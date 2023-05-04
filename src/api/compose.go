package api

/*
	to satisfy the use who need docker-compose, kisara provides
	the api of convert docker-compose.yaml to kisara service config format
	also supports convert reversely
*/

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/types"
)

// convert docker-compose file into kisara config
func ConvertFromCompose(compose_file *types.DockerComposeFile, random_network bool) (*types.ServiceConfig, error) {
	networks := compose_file.Networks
	services := compose_file.Services

	result_config := types.ServiceConfig{}
	result_config.NetworkCount = len(networks)
	result_config_networks := []types.ServiceConfigContainerNetwork{}

	if result_config.NetworkCount > 26 {
		return nil, errors.New("network count should less then 27")
	}

	generate_network_name := func() string {
		return string([]rune{rune('A' + len(result_config_networks))})
	}

	if result_config.NetworkCount == 0 {
		return nil, fmt.Errorf("shoud contains at least one network")
	}

	// map docker-compose network name to kisara network
	network_map := make(map[string]string)

	if result_config.NetworkCount != 0 {
		for name, network := range networks {
			if !network.IPAM.Internal {
				return nil, fmt.Errorf("network %v shoud be internal", name)
			}

			if !network.IPAM.Attachable {
				return nil, fmt.Errorf("network %v shoud be attachable", name)
			}

			if random_network {
				if network_map[name] == "" {
					result_network := types.ServiceConfigContainerNetwork{
						Network:    generate_network_name(),
						RandomCIDR: true,
					}
					result_config_networks = append(result_config_networks, result_network)
					network_map[name] = result_network.Network
				}
			} else {
				return nil, fmt.Errorf("only supports random CIDR currently")
			}
		}
	}

	containers := []types.ServiceConfigContainer{}

	for service_name, service := range services {
		if service.Build != "" {
			return nil, fmt.Errorf("do not support build image in docker-compose, found build config in service %v", service_name)
		}

		if len(service.Networks) == 0 {
			return nil, fmt.Errorf("every service %v should connect to at least 1 network", service_name)
		}

		if service.Image == "" {
			return nil, fmt.Errorf("service %v has no image config", service_name)
		}

		container := types.ServiceConfigContainer{
			Image: service.Image,
		}

		for network_name := range service.Networks {
			// check if network_name is declared in networks
			kisara_network := network_map[network_name]
			if kisara_network == "" {
				return nil, fmt.Errorf("service %v try to connect to a undeclared network %v", service_name, network_name)
			}

			container.Networks = append(container.Networks, types.ServiceConfigContainerNetwork{
				Network:    kisara_network,
				RandomCIDR: true,
			})
		}

		ports := service.Ports
		for _, port := range ports {
			port_parts := strings.Split(port, ":")
			if len(port_parts) > 2 {
				return nil, fmt.Errorf("service %v's port %v format error", service_name, port)
			}

			if len(port_parts) == 0 {
				return nil, fmt.Errorf("service %v has a empty port", service_name)
			}

			lport := 0
			var err error

			if len(port_parts) == 1 {
				lport, err = strconv.Atoi(port_parts[0])
				if err != nil {
					return nil, fmt.Errorf("service %v has a wrong port %v", service_name, port_parts[0])
				}
			} else {
				lport, err = strconv.Atoi(port_parts[1])
				if err != nil {
					return nil, fmt.Errorf("service %v has a wrong port %v", service_name, port_parts[1])
				}
			}

			container.Ports = append(container.Ports, types.ServiceConfigContainerPortMapping{
				Port:     lport,
				Protocol: "tcp",
			})
		}

		result_config.ContainerCount++
		containers = append(containers, container)
	}

	result_config.Containers = containers

	return &result_config, nil
}

// convert docker-compose file text into kisara config
func ConvertFromComposeText(text string, random_network bool) (*types.ServiceConfig, error) {
	compose := &types.DockerComposeFile{}
	err := compose.FromYaml(text)
	if err != nil {
		return nil, err
	}

	return ConvertFromCompose(compose, random_network)
}

// convert kisara service config into docker-compose
func ConvertToCompose(kisara_config *types.ServiceConfig) (*types.DockerComposeFile, error) {
	compose := &types.DockerComposeFile{}

	// check how many networks kisara config has
	network_map := make(map[string]types.DockerComposeFileNetwork)

	services := make(map[string]types.DockerComposeFileService)

	for _, container := range kisara_config.Containers {
		// check how many network the container should connect to
		service := types.DockerComposeFileService{
			Image:    container.Image,
			Networks: make(map[string]types.DockerComposeFileServiceNetwork),
		}

		for _, port := range container.Ports {
			service.Ports = append(service.Ports, strconv.Itoa(port.Port))
		}

		for _, network := range container.Networks {
			if _, ok := network_map[network.Network]; !ok {
				network_map[network.Network] = types.DockerComposeFileNetwork{
					IPAM: types.DockerComposeFileNetworkIPAM{
						Driver:     "overlay",
						Attachable: true,
						Internal:   true,
						Config:     []types.DockerComposeFileNetworkIPAMConfig{},
					},
				}
			}

			service.Networks[network.Network] = types.DockerComposeFileServiceNetwork{}
		}

		services[helper.RandomStr(8)] = service
	}

	compose.Networks = network_map
	compose.Services = services

	return compose, nil
}

func ConvertToComposeText(kisara_config *types.ServiceConfig) (string, error) {
	compose, err := ConvertToCompose(kisara_config)
	if err != nil {
		return "", err
	}

	text := compose.ToYaml()

	return text, nil
}
