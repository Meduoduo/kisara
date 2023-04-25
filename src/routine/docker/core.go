package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Yeuoly/Takina/src/api"
	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/routine/db"
	log "github.com/Yeuoly/kisara/src/routine/log"
	kisara_types "github.com/Yeuoly/kisara/src/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	uuid "github.com/satori/go.uuid"
)

type Docker struct {
	Client *client.Client
	Ctx    *context.Context
}

type portMapping struct {
	ContainerInnerPort int    `json:"container_inner_port"`
	Laddr              string `json:"laddr"`
	Lport              int    `json:"lport"`
	Rport              int    `json:"rport"`
	Raddr              string `json:"raddress"`
	Protocol           string `json:"protocol"`
}

var docker_dns string

type containerMonitor struct {
	ContainerId string
	CPUUsage    uint64
	MemUsage    uint64
	CPUTotal    uint64
	MemTotal    uint64
	CPUPer      float64
	MemPer      float64
}

var containerMonitors sync.Map

func setMonitor(container_id string, stats containerMonitor) {
	containerMonitors.Store(container_id, stats)
}

func getMonitor(container_id string) (containerMonitor, bool) {
	v, ok := containerMonitors.Load(container_id)
	if !ok {
		return containerMonitor{}, false
	}
	return v.(containerMonitor), true
}

func delMonitor(container_id string) {
	containerMonitors.Delete(container_id)
}

func attachMonitor(container_id string) {
	cli := NewDocker()
	stats, err := cli.Client.ContainerStats(*cli.Ctx, container_id, true)
	if err != nil {
		log.Warn("[docker] attach monitor failed %s", err.Error())
		return
	}

	defer stats.Body.Close()
	defer func() {
		delMonitor(container_id)
	}()

	dec := json.NewDecoder(stats.Body)
	for {
		var v *types.StatsJSON
		if err := dec.Decode(&v); err != nil {
			if err == io.EOF {
				break
			}

			log.Warn("[docker] attach monitor failed %s", err.Error())
			return
		}

		if v == nil {
			continue
		}

		cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(v.CPUStats.SystemUsage - v.PreCPUStats.SystemUsage)
		cpuPercent := (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0

		setMonitor(container_id, containerMonitor{
			ContainerId: container_id,
			CPUUsage:    v.CPUStats.CPUUsage.TotalUsage,
			MemUsage:    v.MemoryStats.Usage,
			CPUTotal:    v.CPUStats.SystemUsage,
			MemTotal:    v.MemoryStats.Limit,
			CPUPer:      cpuPercent,
			MemPer:      float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100.0,
		})
	}
}

func InitDocker() {
	docker_dns = helper.GetConfigString("kisara.dns")
	if docker_dns == "" {
		log.Panic("[docker] docker dns not set")
	}

	//关闭所有处于运行中的docker，并删除镜像
	c := NewDocker()
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		log.Panic("[docker] docker start failed")
	}

	defer cli.Close()

	containers, err := cli.ContainerList(*c.Ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		log.Panic("[docker] init docker failed")
	}

	for _, container := range containers {
		// check if container belongs to irina
		if container.Labels["irina"] == "true" {
			err := c.StopContainer(container.ID)
			if err != nil {
				log.Error("[docker] stop docker container failed")
			}
		} else {
			if !strings.Contains(container.Status, "Exit") {
				// attach monitor
				go attachMonitor(container.ID)
			}
		}
	}

	networks, err := c.Client.NetworkList(*c.Ctx, types.NetworkListOptions{})
	if err != nil {
		log.Panic("[docker] init docker failed")
	}

	kisara_networks := make([]kisara_types.Network, 0)

	for _, network := range networks {
		if len(network.IPAM.Config) == 0 {
			continue
		}

		kisara_network := kisara_types.Network{
			Name:     network.Name,
			Id:       network.ID,
			Subnet:   network.IPAM.Config[0].Subnet,
			Internal: network.Internal,
			Driver:   network.Driver,
			Scope:    network.Scope,
		}

		kisara_networks = append(kisara_networks, kisara_network)
	}

	go callOnDockerDaemonStartHooks(c, kisara_networks)

	log.Info("[docker] init docker finished")
}

var global_docker_instance *Docker

func NewDocker() *Docker {
	if global_docker_instance != nil {
		if _, err := global_docker_instance.Client.Ping(context.Background()); err == nil {
			return global_docker_instance
		}
	}
	c := Docker{}
	ctx := helper.GetContext()
	c.Ctx = &ctx
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Warn("[docker] docker start failed %s", err.Error())
		return nil
	}
	c.Client = cli
	global_docker_instance = &c
	return global_docker_instance
}

func (c *Docker) Stop() {
	c.Client.Close()
}

func (c *Docker) PullImage(image_name string, event_callback func(message string)) (*kisara_types.Image, error) {
	image := kisara_types.Image{
		Name: image_name,
	}

	reader, err := c.Client.ImagePull(*c.Ctx, image_name, types.ImagePullOptions{})

	if err != nil || reader == nil {
		return nil, err
	}

	for {
		buf := make([]byte, 1024)
		n, err := reader.Read(buf)

		if err == nil && event_callback != nil {
			event := string(buf[0:n])
			event_callback(event)
		}

		if err == io.EOF || n == 0 {
			break
		}

		if err != nil {
			return nil, err
		}
	}

	return &image, nil
}

func (c *Docker) CreateContainer(image *kisara_types.Image, uid int, port_protocol string, subnet_names []string, module string, env_mount ...map[string]string) (*kisara_types.Container, error) {
	log.Info("[docker] start launch container:" + image.Name)

	// check if subnet exists
	endpoints := make(map[string]*network.EndpointSettings)
	for _, subnet_name := range subnet_names {
		subnet_instance, err := c.GetNetworkByName(subnet_name)
		if err != nil {
			return nil, err
		}

		endpoints[subnet_name] = &network.EndpointSettings{
			NetworkID: subnet_instance.Id,
		}
	}

	default_network_name := "bridge"
	if len(subnet_names) > 0 {
		default_network_name = subnet_names[0]
	}

	/*
		date: 2022/11/19 author: Yeuoly
		to forward compatibility, we do not change the default port protocol
		but at the last version, docker.ContainerCreate only support one port protocol
		therefore, '80/tcp' will be changed to '80/tcp,123/tcp'
	*/

	//create env
	envs := []string{}
	if len(env_mount) > 0 {
		for k, v := range env_mount[0] {
			envs = append(envs, k+"="+v)
		}
	}

	mounts := []mount.Mount{}
	if len(env_mount) > 1 {
		for k, v := range env_mount[1] {
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: k,
				Target: v,
				// set max mount size to 100MB
				//Options: []string{"size=1g"},
				TmpfsOptions: &mount.TmpfsOptions{
					SizeBytes: 100 * 1024 * 1024,
				},
			})
		}
	}

	uuid := uuid.NewV4().String()
	resp, err := c.Client.ContainerCreate(
		*c.Ctx,
		&container.Config{
			Image:        image.Name,
			User:         image.User,
			Tty:          false,
			AttachStdin:  true,
			AttachStdout: true,
			Env:          envs,
			Labels: map[string]string{
				"owner_uid": strconv.Itoa(uid),
				"uuid":      uuid,
				"module":    module,
				"irina":     "true",
			},
		},
		&container.HostConfig{
			NetworkMode: container.NetworkMode(default_network_name),
			Mounts:      mounts,
			Resources: container.Resources{
				//set max memory to 2GB
				Memory: 2 * 1024 * 1024 * 1024,
				//set max cpu to 1 core
				NanoCPUs: 1 * 1000 * 1000 * 1000,
				//set max disk to 5G
				BlkioWeight: 500,
			},
			DNS: []string{docker_dns},
		},
		&network.NetworkingConfig{
			EndpointsConfig: endpoints,
		}, nil, uuid,
	)

	if err != nil {
		return nil, err
	}

	remove_container := func() {
		err := c.Client.ContainerRemove(*c.Ctx, resp.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			log.Warn("[docker] remove container error: " + err.Error())
		}
	}

	err = c.Client.ContainerStart(*c.Ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		remove_container()
		log.Warn("[docker] start container error: " + err.Error())
		return nil, err
	}

	stop_container := func() {
		err := c.Client.ContainerStop(*c.Ctx, resp.ID, container.StopOptions{})
		if err != nil {
			log.Warn("[docker] stop container error: " + err.Error())
		}
	}

	kisara_container := kisara_types.Container{
		Id:    resp.ID,
		Image: image.Name,
		Owner: uid,
		Time:  int(time.Now().Unix()),
		Uuid:  uuid,
	}

	// inspect container to get ip
	inspect, err := c.Client.ContainerInspect(*c.Ctx, resp.ID)
	if err != nil {
		stop_container()
		remove_container()
		log.Warn("[docker] inspect container error: " + err.Error())
		return nil, err
	}

	// get at least one ip
	container_default_ip := ""
	for _, network := range inspect.NetworkSettings.Networks {
		container_default_ip = network.IPAddress
	}

	// parse port protocol
	host_port := ""
	port_protocols := strings.Split(strings.TrimSpace(port_protocol), ",")
	port_mappings := make([]portMapping, 0)

	release := func() {
		for _, port_mapping := range port_mappings {
			if port_mapping.Rport != 0 {
				_, err := api.StopProxy(port_mapping.Laddr, port_mapping.Lport)
				if err != nil {
					log.Warn("[docker] stop proxy %s:%d error: %s", port_mapping.Laddr, port_mapping.Lport, err.Error())
				}
			}
		}
	}

	for i, port_protocol := range port_protocols {
		if len(port_protocol) == 0 {
			continue
		}

		//request launch proxy, protocol_port likes 80/tcp
		protocol_ports := strings.Split(port_protocol, "/")
		if len(protocol_ports) != 2 {
			release()
			stop_container()
			remove_container()
			return nil, errors.New("protocol_port error")
		}

		protocol := protocol_ports[1]
		port, err := strconv.Atoi(protocol_ports[0])
		if err != nil {
			release()
			stop_container()
			remove_container()
			return nil, errors.New("protocol_port format with port error")
		}

		port_mappings = append(port_mappings, portMapping{})
		port_mappings[i].Protocol = protocol_ports[1]
		resp, err := api.StartProxy(container_default_ip, port, protocol)
		if err != nil {
			release()
			stop_container()
			remove_container()
			return nil, err
		}

		r_addr := resp.Proxy.Raddr
		r_port := resp.Proxy.Rport
		host_port += fmt.Sprintf("%s/%s:%d->%s:%d,", protocol, container_default_ip, port, r_addr, r_port)

		log.Info("[docker] start proxy %s:%d -> %s:%d", container_default_ip, port, r_addr, r_port)

		port_mappings[i].ContainerInnerPort, _ = strconv.Atoi(protocol_ports[0])
		port_mappings[i].Laddr = container_default_ip
		port_mappings[i].Lport = port
		port_mappings[i].Rport = r_port
		port_mappings[i].Raddr = r_addr
	}

	kisara_container.HostPort = host_port

	port_map_str, _ := json.Marshal(port_mappings)
	labels := map[string]string{
		"owner_uid": strconv.Itoa(uid),
		"uuid":      uuid,
		"module":    module,
		"irina":     "true",
		"port_map":  string(port_map_str),
		"host_port": host_port,
	}
	labels_str, _ := json.Marshal(labels)

	// create db record
	db_container := &kisara_types.DBContainer{
		ContainerName: kisara_container.Uuid,
		ContainerId:   kisara_container.Id,
		Image:         kisara_container.Image,
		Uid:           kisara_container.Owner,
		Labels:        string(labels_str),
	}

	err = db.CreateGeneric(db_container)
	if err != nil {
		log.Warn("[docker] create db record error: " + err.Error())
		release()
		stop_container()
		remove_container()
		return nil, err
	}

	log.Info("[docker] launch docker successfully: " + kisara_container.Id)

	go attachMonitor(kisara_container.Id)
	go callOnContainerLaunchHooks(c, kisara_container)

	return &kisara_container, nil
}

func (c *Docker) CheckImageExist(image_name string) bool {
	images, err := c.Client.ImageList(*c.Ctx, types.ImageListOptions{})
	if err != nil {
		log.Warn("[docker] list images error: " + err.Error())
		return false
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == image_name {
				return true
			}
		}
	}

	return false
}

func (c *Docker) LaunchTargetMachine(image_name string, port_protocol string, subnet_name string, uid int, module string) (*kisara_types.Container, error) {
	image := &kisara_types.Image{
		Name: image_name,
		User: "root",
	}

	container, err := c.CreateContainer(image, uid, port_protocol, []string{subnet_name}, module)
	if err != nil {
		log.Warn("[docker] create container failed: " + err.Error())
		return nil, err
	}

	log.Info("[docker] launch target machine successfully: " + container.Id)

	return container, nil
}

func (c *Docker) LaunchContainer(image_name string, uid int, port_protocol string, subnet_name string, module string, env_mount ...map[string]string) (*kisara_types.Container, error) {
	image := &kisara_types.Image{
		Name: image_name,
		User: "root",
	}

	container, err := c.CreateContainer(image, uid, port_protocol, []string{subnet_name}, module, env_mount...)
	if err != nil {
		log.Warn("[docker] create container failed: " + err.Error())
		return nil, err
	}

	log.Info("[docker] launch target machine successfully: " + container.Id)

	return container, nil
}

func (c *Docker) LaunchAWD(image_name string, port_protocols string, uid int, subnet_name string, env map[string]string) (*kisara_types.Container, error) {
	image := &kisara_types.Image{
		Name: image_name,
		User: "root",
	}

	//创建容器并留下记录
	container, err := c.CreateContainer(image, uid, port_protocols, []string{subnet_name}, "awd", env)
	if err != nil {
		log.Warn("[docker] create AWD container failed: " + err.Error())
		return nil, err
	}

	log.Info("[docker] launch AWD successfully: " + container.Id)

	return container, nil
}

func (c *Docker) LaunchServiceContainer(image_name string, port_protocols string, uid int, subnet_names []string, env map[string]string) (*kisara_types.Container, error) {
	image := &kisara_types.Image{
		Name: image_name,
		User: "root",
	}

	//创建容器并留下记录
	container, err := c.CreateContainer(image, uid, port_protocols, subnet_names, "service", env)
	if err != nil {
		log.Warn("[docker] create service container failed: " + err.Error())
		return nil, err
	}

	log.Info("[docker] launch service container successfully: " + container.Id)

	return container, nil
}

func (c *Docker) StopContainer(id string) error {
	log.Info("[docker] stop conatiner: " + id)
	//get container labels
	inspect_container, err := c.Client.ContainerInspect(*c.Ctx, id)
	owner_id, _ := strconv.Atoi(inspect_container.Config.Labels["owner_uid"])
	kisara_container := kisara_types.Container{
		Id:    id,
		Image: inspect_container.Config.Image,
		Uuid:  inspect_container.Config.Labels["uuid"],
		Owner: owner_id,
	}
	if err == nil {
		// get db container
		container, err := db.GetGenericOne[kisara_types.DBContainer](
			db.GenericEqual("container_id", id),
		)

		if err != nil {
			return errors.New("could not find container")
		}

		//delete proxy
		labels_str := container.Labels
		labels := make(map[string]string)

		if err := json.Unmarshal([]byte(labels_str), &labels); err != nil {
			return errors.New("could not unmarshal labels in db")
		}

		kisara_container.HostPort = labels["host_port"]
		kisara_container.Labels = labels

		port_map := labels["port_map"]
		if port_map != "" {
			var port_map_map []portMapping
			err = json.Unmarshal([]byte(port_map), &port_map_map)
			if err != nil {
				log.Warn("[docker] unmarshal port map failed: " + err.Error())
			} else {
				for _, port := range port_map_map {
					_, err := api.StopProxy(port.Laddr, port.Lport)
					if err != nil {
						log.Warn("[docker] delete proxy failed: " + err.Error())
					} else {
						log.Info("[docker] delete proxy %s:%d successfully", port.Laddr, port.Lport)
					}
				}
			}
		}
	} else {
		log.Warn("[docker] inspect container failed: " + err.Error())
		return err
	}

	err = c.Client.ContainerStop(*c.Ctx, id, container.StopOptions{})
	if err != nil {
		return nil
	}
	err = c.Client.ContainerRemove(*c.Ctx, id, types.ContainerRemoveOptions{})
	if err == nil {
		callOnContainerStopHooks(c, kisara_container)
	}

	return err
}

func (c *Docker) RemoveContainer(id string) error {
	log.Info("[docker] remove conatiner: " + id)
	err := c.Client.ContainerRemove(*c.Ctx, id, types.ContainerRemoveOptions{})
	return err
}

func (c *Docker) Exec(container_id string, cmd string) error {
	exec, err := c.Client.ContainerExecCreate(*c.Ctx, container_id, types.ExecConfig{
		AttachStdin:  false,
		AttachStderr: true,
		AttachStdout: true,
		Tty:          false,
		User:         "root",
		Cmd:          []string{"sh", "-c", cmd},
	})
	if err != nil {
		return err
	}

	resp, err := c.Client.ContainerExecAttach(*c.Ctx, exec.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return err
	}

	_, err = io.ReadAll(resp.Reader)

	if err != nil {
		return err
	}

	return nil
}

func (c *Docker) ListContainer() (*[]*kisara_types.Container, error) {
	containers, err := c.Client.ContainerList(*c.Ctx, types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		return nil, err
	}

	var container_list []*kisara_types.Container
	for _, container := range containers {
		labels := make(map[string]string)

		db_container, err := db.GetGenericOne[kisara_types.DBContainer](
			db.GenericEqual("container_id", container.ID),
		)

		if err == nil {
			labels_str := db_container.Labels
			json.Unmarshal([]byte(labels_str), &labels)
		}

		owner_uid, _ := strconv.Atoi(container.Labels["owner_uid"])
		container_list = append(container_list, &kisara_types.Container{
			Id:       container.ID,
			Image:    container.Image,
			Owner:    owner_uid,
			Time:     int(container.Created),
			Uuid:     labels["uuid"],
			HostPort: labels["host_port"],
			Status:   container.Status,
		})
	}

	return &container_list, nil
}

func (c *Docker) ListImage() (*[]*kisara_types.Image, error) {
	images, err := c.Client.ImageList(*c.Ctx, types.ImageListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	var image_list []*kisara_types.Image
	for _, image := range images {
		current_image := &kisara_types.Image{}
		current_image.Uuid = image.ID
		if image.RepoTags != nil {
			current_image.Name = image.RepoTags[0]
		} else {
			current_image.Name = image.ID
		}
		current_image.VirtualSize = image.VirtualSize
		image_list = append(image_list, current_image)
	}

	return &image_list, nil
}

func (c *Docker) DeleteImage(uuid string) error {
	_, err := c.Client.ImageRemove(*c.Ctx, uuid, types.ImageRemoveOptions{
		Force: true,
	})
	if err != nil {
		return err
	}
	return nil
}

/*
InspectContainer will insepct to docker container and return some
information about the container like host_port, container status, etc.
*/
func (c *Docker) InspectContainer(container_id string, has_state ...bool) (*kisara_types.Container, error) {
	container, err := c.Client.ContainerInspect(*c.Ctx, container_id)
	if err != nil {
		return nil, err
	}

	var cpu_usage float64
	var memory_usage float64

	if len(has_state) > 0 && has_state[0] {
		stats, ok := getMonitor(container_id)
		if ok {
			cpu_usage = stats.CPUPer
			memory_usage = stats.MemPer
		}
	}

	db_container, err := db.GetGenericOne[kisara_types.DBContainer](
		db.GenericEqual("container_id", container_id),
	)

	if err != nil {
		return nil, errors.New("unable to find container in kisara db")
	}

	labels_str := db_container.Labels
	labels := make(map[string]string)

	if err := json.Unmarshal([]byte(labels_str), &labels); err != nil {
		return nil, errors.New("could not unmarshal labels in db")
	}

	ret := &kisara_types.Container{
		Id:       container.ID,
		HostPort: container.Config.Labels["host_port"],
		Status:   container.Config.Labels["status"],
		Labels:   labels,
		// cpu usage
		// memory usage
		CPUUsage: cpu_usage,
		MemUsage: memory_usage,
	}

	return ret, nil
}

/*
Get container Number
*/
func (c *Docker) GetContainerNumber() (int, error) {
	containers, err := c.Client.ContainerList(*c.Ctx, types.ContainerListOptions{})
	if err != nil {
		return 0, err
	}
	return len(containers), nil
}
