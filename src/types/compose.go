package types

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type DockerComposeFile struct {
	Services map[string]DockerComposeFileService `yaml:"services"`
	Networks map[string]DockerComposeFileNetwork `yaml:"networks"`
	Volumes  map[string]struct{}                 `yaml:"volumes"`
}

type DockerComposeFileNetwork struct {
	IPAM DockerComposeFileNetworkIPAM `yaml:"ipam"`
}

type DockerComposeFileNetworkIPAM struct {
	Driver     string                               `yaml:"driver"`
	Attachable bool                                 `yaml:"attachable"`
	Internal   bool                                 `yaml:"internal"`
	Config     []DockerComposeFileNetworkIPAMConfig `yaml:"config"`
}

type DockerComposeFileNetworkIPAMConfig struct {
	Subnet string `yaml:"subnet"`
}

type DockerComposeFileService struct {
	Image    string                                     `yaml:"image"`
	Build    string                                     `yaml:"build"`
	Networks map[string]DockerComposeFileServiceNetwork `yaml:"networks"`
	Ports    []string                                   `yaml:"ports"`
}

type DockerComposeFileServiceNetwork struct {
	Ipv4Address string `yaml:"ipv4_address"`
	Ipv6Address string `yaml:"ipv6_address"`
}

func (c *DockerComposeFile) ToYaml() string {
	result, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}

	return string(result)
}

func (c *DockerComposeFile) FromYaml(text string) error {
	text = strings.ReplaceAll(text, "\t", "  ")
	return yaml.Unmarshal([]byte(text), c)
}
