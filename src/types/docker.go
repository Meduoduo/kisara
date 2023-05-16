package types

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Image struct {
	Id           int    `json:"id"`
	Uuid         string `json:"uuid"`
	Name         string `json:"name"`
	User         string `json:"user"`
	LastUpdate   int    `json:"last_update"`
	PortProtocol string `json:"port_protocol"`
	VirtualSize  int64  `json:"virtual_size"`
}

type Container struct {
	Id       string            `json:"id"`
	Image    string            `json:"image"`
	Uuid     string            `json:"uuid"`
	Time     int               `json:"time"`
	Owner    int               `json:"owner"`
	HostPort string            `json:"host_port"`
	Labels   map[string]string `json:"labels"`
	Status   string            `json:"status"`
	CPUUsage float64           `json:"cpu_usage"`
	MemUsage float64           `json:"mem_usage"`
	Networks []Network         `json:"networks"`
}

type Network struct {
	Id       string `json:"id"`
	Subnet   string `json:"subnet"`
	Name     string `json:"name"`
	Internal bool   `json:"internal"`
	Driver   string `json:"driver"`
	Scope    string `json:"scope"`
}

// ServiceFlag Contains the flag kisara generated for the service, it's not the part of service config
type ServiceFlag struct {
	FlagUuid string `json:"flag_uuid"`
	Flag     string `json:"flag"`
}

// Service is the service kisara generated for the user
type Service struct {
	Id         string        `json:"id"`
	Name       string        `json:"name"`
	Containers []Container   `json:"containers"`
	Networks   []Network     `json:"networks"`
	Flags      []ServiceFlag `json:"flags"`
	Status     string        `json:"status"`
}

const (
	SERVICE_STATUS_RUNNING = "running"
)

// Service is not the service in docker compose, it is the service config in kisara
type KisaraService struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       int    `json:"owner"`
	Config      string `json:"config"` // json string of service config
}

type ServiceConfigContainerPortMapping struct {
	Port     int    `json:"port" yaml:"port"`         // port in container to be mapped
	Protocol string `json:"protocol" yaml:"protocol"` // protocol of port
}

type ServiceConfigContainerNetwork struct {
	Network    string `json:"network" yaml:"network"`        // network to be used, if random_network is true, this field should be a string of network name
	RandomCIDR bool   `json:"random_cidr" yaml:"RandomCIDR"` // whether to generate random container
}

type ServiceConfigContainer struct {
	Image    string                              `json:"image" yaml:"image"`
	Ports    []ServiceConfigContainerPortMapping `json:"ports" yaml:"ports"`
	Networks []ServiceConfigContainerNetwork     `json:"networks" yaml:"networks"`
	Flags    []ServiceConfigContainerFlag        `json:"flags" yaml:"flags"`
	Env      map[string]string                   `json:"env" yaml:"env"`
}

type ServiceConfigContainerFlag struct {
	FlagCommand string `json:"flag_command" yaml:"flag_command"`
	FlagScore   int    `json:"flag_score" yaml:"flag_score"`
	FlagUuid    string `json:"flag_uuid" yaml:"flag_uuid"` // uuid of flag
}

type ServiceConfig struct {
	Containers     []ServiceConfigContainer `json:"containers" yaml:"containers"`
	TotalScore     int                      `json:"total_score" yaml:"total_score"`
	NetworkCount   int                      `json:"network_count" yaml:"network_count"`
	ContainerCount int                      `json:"container_count" yaml:"container_count"`
}

func (c *KisaraService) GetConfig() (ServiceConfig, error) {
	var config ServiceConfig
	err := json.Unmarshal([]byte(c.Config), &config)
	// check config format
	if err != nil {
		return config, err
	}

	networks := make(map[string]bool)

	total_score := config.TotalScore
	for _, container := range config.Containers {
		for _, network := range container.Networks {
			if len(network.Network) == 0 {
				return config, errors.New("network cannot be empty")
			}
			if network.RandomCIDR {
				if len(network.Network) != 1 {
					return config, errors.New("random CIDR network should be one character like 'A' or 'B'")
				}
				if !strings.Contains("ABCDEFGHIJKLMNOPQRSTUVWXYZ", network.Network) {
					return config, errors.New("random CIDR network should be one character like 'A' or 'B'")
				}
			} else {
				return config, errors.New("currently only support random CIDR network")
			}
			networks[network.Network] = true
		}

		for _, flag := range container.Flags {
			total_score -= flag.FlagScore
		}
	}

	if total_score != 0 {
		return config, errors.New("total score of service does not match the sum of container scores")
	}

	if len(networks) != config.NetworkCount {
		return config, errors.New("network count does not match the number of networks")
	}

	if len(config.Containers) != config.ContainerCount {
		return config, errors.New("container count does not match the number of containers")
	}

	return config, nil
}

func (c *ServiceConfig) RandomCIDRCount() int {
	count := 0
	for _, container := range c.Containers {
		for _, network := range container.Networks {
			if network.RandomCIDR {
				count++
			}
		}
	}
	return count
}

func (c *ServiceConfig) GetNetworks() []ServiceConfigContainerNetwork {
	networks := make(map[string]ServiceConfigContainerNetwork)
	for _, container := range c.Containers {
		for _, network := range container.Networks {
			networks[network.Network] = network
		}
	}
	networks_list := make([]ServiceConfigContainerNetwork, 0)
	for _, network := range networks {
		networks_list = append(networks_list, network)
	}
	return networks_list
}

func (c *ServiceConfig) GetFlagCount() int {
	count := 0
	for _, container := range c.Containers {
		count += len(container.Flags)
	}
	return count
}

func (c *ServiceConfigContainer) GetPortProtocol() map[int]string {
	port_protocol := make(map[int]string)
	for _, port := range c.Ports {
		port_protocol[port.Port] = port.Protocol
	}
	return port_protocol
}

func (c *ServiceConfigContainer) GetPortProtocolText() string {
	port_protocols := []string{}
	for port, protocol := range c.GetPortProtocol() {
		port_protocols = append(port_protocols, strconv.Itoa(port)+"/"+protocol)
	}
	return strings.Join(port_protocols, ",")
}

func (c *ServiceConfig) ToYaml() string {
	result, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}

	return string(result)
}

func (c *ServiceConfig) FromYaml(text_config string) error {
	return yaml.Unmarshal([]byte(text_config), c)
}
