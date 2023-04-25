package docker

import (
	"errors"
	"strings"

	log "github.com/Yeuoly/kisara/src/routine/log"
	"github.com/Yeuoly/kisara/src/types"
	uuid "github.com/satori/go.uuid"
)

// create a service from a service config
func (c *Docker) CreateService(service_config types.KisaraService) (*types.Service, error) {
	config, err := service_config.GetConfig()
	if err != nil {
		return nil, err
	}

	networks := config.GetNetworks()
	if len(networks) > 4 {
		return nil, errors.New("at most 4 networks are allowed")
	}

	if len(networks) == 0 {
		return nil, errors.New("at least 1 network is required")
	}

	if config.ContainerCount == 0 {
		return nil, errors.New("container count cannot be 0")
	}

	if config.ContainerCount > 16 {
		return nil, errors.New("at most 16 containers are allowed")
	}

	// create service
	type network_info struct {
		network      *types.Network
		OriginalName string
	}

	// create networks
	result_networks := make([]*network_info, 0)
	release_networks := func() {
		for _, network := range result_networks {
			if strings.HasPrefix(network.network.Name, "kisara_") {
				err = c.ReleaseCIDRNetwork(network.network.Name[7:])
			} else {
				err = c.DeleteNetwork(network.network.Id)
			}
			if err != nil {
				log.Warn("[service] release network failed: %s", err.Error())
			}
		}
	}

	for _, network := range networks {
		if network.RandomCIDR {
			net, err := c.CreateRandomCIDRNetwork(true, "overlay")
			if err != nil {
				release_networks()
				return nil, err
			}
			result_networks = append(result_networks, &network_info{
				network:      net,
				OriginalName: network.Network,
			})
		} else {
			net, err := c.CreateNetwork(uuid.NewV4().String(), network.Network, true, "overlay")
			if err != nil {
				release_networks()
				return nil, err
			}
			result_networks = append(result_networks, &network_info{
				network:      net,
				OriginalName: network.Network,
			})
		}
	}

	// create containers
	result_containers := make([]types.Container, 0)
	release_containers := func() {
		for _, container := range result_containers {
			err = c.StopContainer(container.Id)
			if err != nil {
				log.Warn("[service] release container failed: %s", err.Error())
			}
		}
		release_networks()
	}

	flags := make([]types.ServiceFlag, 0)

	for i := 0; i < len(config.Containers); i++ {
		container_config := config.Containers[i]
		// find networks the container should be connected to
		networks := make([]*types.Network, 0)
		for _, network := range result_networks {
			for _, container_network := range container_config.Networks {
				if container_network.Network == network.OriginalName {
					networks = append(networks, network.network)
				}
			}
		}

		network_names := make([]string, 0)
		for _, network := range networks {
			network_names = append(network_names, network.Name)
		}

		container, err := c.LaunchServiceContainer(container_config.Image, container_config.GetPortProtocolText(), service_config.Owner, network_names, container_config.Env)
		if err != nil {
			release_containers()
			return nil, err
		}

		result_containers = append(result_containers, *container)

		// execute flag command
		for _, flag := range container_config.Flags {
			flag_text := `kisara{` + uuid.NewV4().String() + `}`
			flag_command := strings.Replace(flag.FlagCommand, "$flag", flag_text, -1)
			err := c.Exec(container.Id, flag_command)
			if err != nil {
				release_containers()
				return nil, err
			}

			flags = append(flags, types.ServiceFlag{
				FlagUuid: flag.FlagUuid,
				Flag:     flag_text,
			})
		}
	}

	networks_result := make([]types.Network, 0)
	for _, network := range result_networks {
		networks_result = append(networks_result, *network.network)
	}

	// create service
	service := &types.Service{
		Containers: result_containers,
		Networks:   networks_result,
		Flags:      flags,
	}

	return service, nil
}

// delete a service
func (c *Docker) DeleteService(service *types.Service) error {
	for _, container := range service.Containers {
		err := c.StopContainer(container.Id)
		if err != nil {
			return err
		}
	}

	for _, network := range service.Networks {
		if strings.HasPrefix(network.Name, "kisara_") {
			err := c.ReleaseCIDRNetwork(network.Name[7:])
			if err != nil {
				return err
			}
		} else {
			err := c.DeleteNetwork(network.Id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
