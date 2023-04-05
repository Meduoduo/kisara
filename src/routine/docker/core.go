package routine

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

	"github.com/Yeuoly/kisara/src/helper"
	log "github.com/Yeuoly/kisara/src/routine/log"
	request "github.com/Yeuoly/kisara/src/routine/request"
	takina "github.com/Yeuoly/kisara/src/routine/takina"
	kisara_types "github.com/Yeuoly/kisara/src/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	uuid "github.com/satori/go.uuid"
)

type Docker struct {
	Client *client.Client
	Ctx    *context.Context
}

type portMapping struct {
	ContainerInnerPort int    `json:"container_inner_port"`
	Lport              int    `json:"lport"`
	Rport              int    `json:"rport"`
	Raddress           string `json:"raddress"`
	Protocol           string `json:"protocol"`
}

var docker_version string
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
	docker_version = "1.38"

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
		panic("[docker] docker start failed")
	}

	defer cli.Close()

	containers, err := cli.ContainerList(*c.Ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		panic("[docker] init docker failed")
	}

	for _, container := range containers {
		// check if container belongs to irina
		if container.Labels["irina"] == "true" {
			err := c.StopContainer(container.ID)
			if err != nil {
				panic("[docker] stop docker container failed")
			}
		} else {
			if !strings.Contains(container.Status, "Exit") {
				// attach monitor
				go attachMonitor(container.ID)
			}
		}
	}

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

func DockerPullImage(cli *Docker, image_name string, event_callback func(message string)) (*kisara_types.Image, error) {
	image := kisara_types.Image{
		Name: image_name,
	}

	reader, err := cli.Client.ImagePull(*cli.Ctx, image_name, types.ImagePullOptions{})

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

// 用于控制器启动子线程拉取镜像
func HandleControllerRequestPullImage(request_id string, image_name string, port_protocol string, user string) {
	docker := NewDocker()

	message_callback := func(message string) {
		request.SetRequestStatusText(request_id, message)
	}
	_, err := DockerPullImage(docker, image_name, message_callback)

	var response struct {
		Res int `json:"res"`
	}

	if err != nil {
		response.Res = -1
	} else {
		response.Res = 0
	}

	text, _ := json.Marshal(response)
	request.FinishRequest(request_id, string(text))

	/*
		TODO: add localstorage record
	*/
}

func (c *Docker) CreateContainer(image *kisara_types.Image, uid int, port_protocol string, subnet_name string, module string, env_mount ...map[string]string) (*kisara_types.Container, error) {
	log.Info("[docker] start launch container:" + image.Name)

	// check if subnet exists
	subnet_instance, err := c.GetNetworkByName(subnet_name)
	if err != nil {
		return nil, err
	}

	subnet_id := subnet_instance.Id

	/*
		date: 2022/11/19 author: Yeuoly
		to forward compatibility, we do not change the default port protocol
		but at the last version, docker.ContainerCreate only support one port protocol
		therefore, '80/tcp' will be changed to '80/tcp,123/tcp'
	*/

	port_protocols := strings.Split(port_protocol, ",")
	port_mappings := make([]portMapping, len(port_protocols))

	release := func() {
		for _, port_mapping := range port_mappings {
			if port_mapping.Rport != 0 {
				takina.TakinaRequestDelProxy("127.0.0.1", port_mapping.Lport)
			}
		}
	}

	for i, port_protocol := range port_protocols {
		port, err := helper.GetAvaliablePort()
		if err != nil {
			release()
			return nil, err
		}

		//request launch proxy, protocol_port likes 80/tcp
		protocol_ports := strings.Split(port_protocol, "/")
		if len(protocol_ports) != 2 {
			release()
			return nil, errors.New("protocol_port error")
		}

		port_mappings[i].Protocol = protocol_ports[1]
		r_addr, r_port, err := takina.TakinaRequestAddProxy("127.0.0.1", port, protocol_ports[1])
		if err != nil {
			release()
			return nil, err
		}

		port_mappings[i].ContainerInnerPort, _ = strconv.Atoi(protocol_ports[0])
		port_mappings[i].Rport = r_port
		port_mappings[i].Lport = port
		port_mappings[i].Raddress = r_addr
	}

	// container label
	host_port := ""
	if len(port_mappings) > 0 {
		host_port = port_mappings[0].Raddress + ":" + strconv.Itoa(port_mappings[0].Rport)
		for i := 1; i < len(port_mappings); i++ {
			host_port += "," + port_mappings[i].Raddress + ":" + strconv.Itoa(port_mappings[i].Rport)
		}
	}

	port_map := make(nat.PortMap)
	for _, port_mapping := range port_mappings {
		port_map[nat.Port(strconv.Itoa(port_mapping.ContainerInnerPort)+"/"+port_mapping.Protocol)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.Itoa(port_mapping.Lport),
			},
		}
	}

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

	json_port_mappings, _ := json.Marshal(port_mappings)

	//networkMode := "none"
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
				"port_map":  string(json_port_mappings),
				"host_port": host_port,
			},
		},
		&container.HostConfig{
			NetworkMode:  container.NetworkMode(subnet_name),
			PortBindings: port_map,
			Mounts:       mounts,
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
			EndpointsConfig: map[string]*network.EndpointSettings{
				"kisara": {
					NetworkID: subnet_id,
				},
			},
		}, nil, uuid,
	)

	if err != nil {
		return nil, err
	}

	container := kisara_types.Container{
		Id:       resp.ID,
		Image:    image.Name,
		Owner:    uid,
		Time:     int(time.Now().Unix()),
		Uuid:     uuid,
		HostPort: host_port,
	}

	err = c.Client.ContainerStart(*c.Ctx, resp.ID, types.ContainerStartOptions{})

	if err != nil {
		log.Warn("[docker] start container error: " + err.Error())
		return nil, err
	}

	log.Info("[docker] launch docker successfully: " + container.Id)

	go attachMonitor(container.Id)

	return &container, nil
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

	container, err := c.CreateContainer(image, uid, port_protocol, subnet_name, module)
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

	container, err := c.CreateContainer(image, uid, port_protocol, subnet_name, module, env_mount...)
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
	container, err := c.CreateContainer(image, uid, port_protocols, subnet_name, "awd", env)
	if err != nil {
		log.Warn("[docker] create AWD container failed: " + err.Error())
		return nil, err
	}

	log.Info("[docker] launch AWD successfully: " + container.Id)

	return container, nil
}

func (c *Docker) StopContainer(id string) error {
	log.Info("[docker] stop conatiner: " + id)
	//get container labels
	_container, err := c.Client.ContainerInspect(*c.Ctx, id)
	if err == nil {
		//delete proxy
		port_map := _container.Config.Labels["port_map"]
		if port_map != "" {
			var port_map_map []portMapping
			err = json.Unmarshal([]byte(port_map), &port_map_map)
			if err != nil {
				log.Warn("[docker] unmarshal port map failed: " + err.Error())
			} else {
				for _, port := range port_map_map {
					err = takina.TakinaRequestDelProxy("127.0.0.1", port.Lport)
					if err != nil {
						log.Warn("[docker] delete proxy failed: " + err.Error())
					}
				}
			}
		}
	}

	err = c.Client.ContainerStop(*c.Ctx, id, container.StopOptions{})
	if err != nil {
		return nil
	}
	err = c.Client.ContainerRemove(*c.Ctx, id, types.ContainerRemoveOptions{})

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

	res, err := io.ReadAll(resp.Reader)
	fmt.Println(string(res))

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
		owner_uid, _ := strconv.Atoi(container.Labels["owner_uid"])
		container_list = append(container_list, &kisara_types.Container{
			Id:       container.ID,
			Image:    container.Image,
			Owner:    owner_uid,
			Time:     int(container.Created),
			Uuid:     container.Labels["uuid"],
			HostPort: container.Labels["host_port"],
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

var docker_sync_lock sync.Mutex
var docker_sync_request_id string

func (c *Docker) syncImage(request_id string, images []string) {
	defer docker_sync_lock.Unlock()

	var response struct {
		Res  int    `json:"res"`
		Data string `json:"data"`
		Err  string `json:"err"`
	}

	message_callback := func(message string) {
		request.SetRequestStatusText(request_id, message)
	}

	exists_image, err := c.ListImage()
	if err != nil {
		response.Res = -1
		response.Err = "拉取镜像列表失败"
		response_text, _ := json.Marshal(response)
		request.FinishRequest(request_id, string(response_text))
		return
	}

	check_exist := func(name string) bool {
		for _, image := range *exists_image {
			if image.Name == name {
				return true
			}
		}
		return false
	}

	for _, image := range images {
		if !check_exist(image) {
			_, err := DockerPullImage(c, image, message_callback)
			if err != nil {
				response.Res = -1
				response.Err = "拉取镜像失败"
				response_text, _ := json.Marshal(response)
				request.FinishRequest(request_id, string(response_text))
				return
			}
		}
	}

	response.Res = 1
	response.Data = ""
	response.Err = ""
	response_text, _ := json.Marshal(response)
	request.FinishRequest(request_id, string(response_text))
}

func (c *Docker) StartSyncImage(images []string) (string, error) {
	if !docker_sync_lock.TryLock() {
		return "", errors.New("docker sync is running")
	}

	request_id := request.CreateNewResponse()
	go c.syncImage(request_id, images)
	docker_sync_request_id = request_id
	return request_id, nil
}

func (c *Docker) CheckSyncStatus() (string, error) {
	if !docker_sync_lock.TryLock() {
		return docker_sync_request_id, nil
	}
	docker_sync_lock.Unlock()
	return "", errors.New("no sync")
}

/* Get message from docker image sync task */
func (c *Docker) GetSyncMessage() string {
	if docker_sync_lock.TryLock() {
		docker_sync_lock.Unlock()
		return ""
	}
	res, ok := request.GetResponse(docker_sync_request_id)
	if !ok {
		return ""
	}
	return res
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

	ret := &kisara_types.Container{
		Id:       container.ID,
		HostPort: container.Config.Labels["host_port"],
		Status:   container.Config.Labels["status"],
		Labels:   container.Config.Labels,
		// cpu usage
		// memory usage
		CPUUsage: cpu_usage,
		MemUsage: memory_usage,
	}

	return ret, nil
}

/*
Create a new docker virtual network
*/
func (c *Docker) CreateNetwork(subnet string, name string, host_join bool) error {
	_, err := c.Client.NetworkCreate(*c.Ctx, name, types.NetworkCreate{
		Driver:         "overlay",
		CheckDuplicate: true,
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet: subnet,
				},
			},
		},
		EnableIPv6: false,
		// if host_join is true, the host will join the network, otherwise the host will not join the network
		Internal:   !host_join,
		Attachable: true,
	})
	if err != nil {
		return err
	}
	return nil
}

/*
Delete a docker virtual network
*/
func (c *Docker) DeleteNetwork(network_id string) error {
	err := c.Client.NetworkRemove(*c.Ctx, network_id)
	if err != nil {
		return err
	}
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
Get container Number
*/
func (c *Docker) GetContainerNumber() (int, error) {
	containers, err := c.Client.ContainerList(*c.Ctx, types.ContainerListOptions{})
	if err != nil {
		return 0, err
	}
	return len(containers), nil
}
