package docker

import (
	"container/list"
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Yeuoly/kisara/src/routine/log"
	kisara_types "github.com/Yeuoly/kisara/src/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

var (
	cidr_pool = list.New()
	cidr_mux  = sync.Mutex{}
)

func requestCIDR() (string, error) {
	cidr_mux.Lock()
	defer cidr_mux.Unlock()

	if cidr_pool.Len() == 0 {
		return "", errors.New("no cidr available")
	}

	cidr := cidr_pool.Front()
	cidr_pool.Remove(cidr)
	return cidr.Value.(string), nil
}

func releaseCIDR(cidr string) {
	cidr_mux.Lock()
	defer cidr_mux.Unlock()
	cidr_pool.PushBack(cidr)
}

/*
Parse a CIDR expression into a list of CIDR
CIDR expression like: 172.[128-255].[0-255].0/24
every number in the expression can be a range or a single number
*/
func ParseCIDRRange(cidr string) ([]string, error) {
	list := make([]string, 0)
	// check if expression contains a range
	left_bracket := strings.Index(cidr, "[")
	right_bracket := strings.Index(cidr, "]")
	if left_bracket == -1 || right_bracket == -1 {
		list = append(list, cidr)
		return list, nil
	}

	// parse the range
	range_str := cidr[left_bracket+1 : right_bracket]
	range_list := strings.Split(range_str, "-")
	if len(range_list) != 2 {
		return list, errors.New("invalid CIDR range")
	}

	// parse the range
	start, err := strconv.Atoi(range_list[0])
	if err != nil {
		return list, err
	}
	end, err := strconv.Atoi(range_list[1])
	if err != nil {
		return list, err
	}

	if start > end || start < 0 || end > 255 {
		return list, errors.New("invalid CIDR range")
	}

	// generate CIDR list
	for i := start; i <= end; i++ {
		temp_list, err := ParseCIDRRange(cidr[:left_bracket] + strconv.Itoa(i) + cidr[right_bracket+1:])
		if err != nil {
			return list, err
		}
		list = append(list, temp_list...)
	}

	return list, nil
}

func InitCIDRPool(cidr_expression string) (int, error) {
	cidr_list, err := ParseCIDRRange(cidr_expression)
	if err != nil {
		return 0, err
	}

	c := NewDocker()
	networks, err := c.ListNetwork()
	if err != nil {
		return 0, err
	}
	for _, cidr := range cidr_list {
		// check if CIDR is used
		for _, network := range networks {
			if network.Name == "kisara_"+strings.Replace(strings.Replace(cidr, "/", "_", -1), ".", "_", -1) {
				c.DeleteNetwork(network.Id)
			}
		}
		cidr_pool.PushBack(cidr)
	}

	return cidr_pool.Len(), nil
}

/*
Create a Random CIDR network
*/
func (c *Docker) CreateRandomCIDRNetwork(internal bool, driver string) (*kisara_types.Network, error) {
	for i := 0; i < 3; i++ {
		cidr, err := requestCIDR()
		if err != nil {
			releaseCIDR(cidr)
			continue
		}

		cidr_name := strings.Replace(cidr, ".", "_", -1)
		cidr_name = "kisara_" + strings.Replace(cidr_name, "/", "_", -1)
		network, err := c.CreateNetwork(cidr, cidr_name, internal, driver)
		if err != nil {
			releaseCIDR(cidr)
			continue
		}

		return network, nil
	}

	return nil, errors.New("no cidr available")
}

/*
Release a CIDR network
*/
func (c *Docker) ReleaseCIDRNetwork(cidr string) error {
	network, err := c.GetNetworkByName("kisara_" + cidr)
	if err != nil {
		return err
	}

	err = c.DeleteNetwork(network.Id)
	if err != nil {
		return err
	}

	releaseCIDR(cidr)
	return nil
}

/*
Get a docker virtual network by name
*/
func (c *Docker) GetNetworkByName(name string) (*kisara_types.Network, error) {
	networks, err := c.ListNetwork()
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		if network.Name == name {
			return &network, nil
		}
	}

	return nil, errors.New("network not found")
}

/*
Create a new docker virtual network
*/
func (c *Docker) CreateNetwork(subnet string, name string, internal bool, driver string) (*kisara_types.Network, error) {
	resp, err := c.Client.NetworkCreate(*c.Ctx, name, types.NetworkCreate{
		Driver:         driver,
		CheckDuplicate: true,
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet: subnet,
				},
			},
		},
		EnableIPv6: false,
		Internal:   internal,
		Attachable: true,
	})
	if err != nil {
		return nil, err
	}

	network := kisara_types.Network{
		Id:       resp.ID,
		Name:     name,
		Subnet:   subnet,
		Internal: internal,
		Driver:   driver,
		Scope:    "swarm",
	}

	// wait for network to be ready
	for i := 0; i < 30; i++ {
		net, _ := c.GetNetworkByName(name)
		if net != nil {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}

	go callOnNetworkCreateHooks(c, network)

	return &network, nil
}

/*
Delete a docker virtual network
*/
func (c *Docker) DeleteNetwork(network_id string) error {
	// inspect network
	net, err := c.Client.NetworkInspect(*c.Ctx, network_id, types.NetworkInspectOptions{})
	if err != nil {
		return err
	}

	if len(net.IPAM.Config) == 0 {
		return errors.New("network does not have subnet")
	}

	network := kisara_types.Network{
		Id:       network_id,
		Name:     net.Name,
		Subnet:   net.IPAM.Config[0].Subnet,
		Internal: net.Internal,
		Driver:   net.Driver,
		Scope:    net.Scope,
	}

	err = callBeforeNetworkRemoveHooks(c, network)
	if err != nil {
		return err
	}

	// check if network is used by any container for 20 seconds
	for i := 0; i < 20; i++ {
		network_check, err := c.Client.NetworkInspect(*c.Ctx, network_id, types.NetworkInspectOptions{})
		if err != nil {
			return err
		}

		if len(network_check.Containers) == 0 {
			break
		}

		log.Warn("[Network] Network %s is still used by containers, waiting...", network_id)
		time.Sleep(1000 * time.Millisecond)
	}

	err = c.Client.NetworkRemove(*c.Ctx, network_id)
	if err != nil {
		return err
	}

	callOnNetworkRemoveHooks(c, network)

	return nil
}

/*
List all docker virtual network
*/
func (c *Docker) ListNetwork() ([]kisara_types.Network, error) {
	networks, err := c.Client.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		log.Warn("List network failed: %s", err.Error())
	}

	var ret []kisara_types.Network
	for _, network := range networks {
		if len(network.IPAM.Config) == 0 {
			continue
		}
		ret = append(ret, kisara_types.Network{
			Id:       network.ID,
			Name:     network.Name,
			Subnet:   network.IPAM.Config[0].Subnet,
			Internal: network.Internal,
			Driver:   network.Driver,
			Scope:    network.Scope,
		})
	}

	return ret, nil
}

/*
Connect a container to a network
*/
func (c *Docker) ConnectContainerToNetwork(container_id string, network_id string) error {
	err := c.Client.NetworkConnect(*c.Ctx, network_id, container_id, nil)
	if err != nil {
		return err
	}
	return nil
}

/*
Disconnect a container from a network
*/
func (c *Docker) DisconnectContainerFromNetwork(container_id string, network_id string) error {
	err := c.Client.NetworkDisconnect(*c.Ctx, network_id, container_id, true)
	if err != nil {
		return err
	}
	return nil
}

type ContainerNetworkInfo struct {
	ContainerId string
	Networks    []struct {
		Network kisara_types.Network
		Ip      string
	}
}

/*
Get a container's network info
*/
func (c *Docker) GetContainerNetwork(container_id string) (ContainerNetworkInfo, error) {
	ret := ContainerNetworkInfo{}
	container, err := c.Client.ContainerInspect(*c.Ctx, container_id)
	if err != nil {
		return ret, err
	}

	network_map := make(map[string]types.NetworkResource)

	for _, network := range container.NetworkSettings.Networks {
		if _, ok := network_map[network.NetworkID]; ok {
			network_resource := network_map[network.NetworkID]

			ret.Networks = append(ret.Networks, struct {
				Network kisara_types.Network
				Ip      string
			}{
				Network: kisara_types.Network{
					Id:       network.NetworkID,
					Name:     network_resource.Name,
					Subnet:   network_resource.IPAM.Config[0].Subnet,
					Internal: network_resource.Internal,
					Driver:   network_resource.Driver,
					Scope:    network_resource.Scope,
				},
				Ip: network.IPAddress,
			})
		}

		docker_network, err := c.Client.NetworkInspect(*c.Ctx, network.NetworkID, types.NetworkInspectOptions{})
		if err != nil {
			return ret, err
		}

		if len(docker_network.IPAM.Config) == 0 {
			continue
		}

		if _, ok := network_map[docker_network.ID]; !ok {
			network_map[docker_network.ID] = docker_network
		}

		ret.Networks = append(ret.Networks, struct {
			Network kisara_types.Network
			Ip      string
		}{
			Network: kisara_types.Network{
				Id:       docker_network.ID,
				Name:     docker_network.Name,
				Subnet:   docker_network.IPAM.Config[0].Subnet,
				Internal: docker_network.Internal,
				Driver:   docker_network.Driver,
				Scope:    docker_network.Scope,
			},
			Ip: network.IPAddress,
		})
	}

	return ret, nil
}
