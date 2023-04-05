package types

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
	Id       string  `json:"id"`
	Image    string  `json:"image"`
	Uuid     string  `json:"uuid"`
	Time     int     `json:"time"`
	Owner    int     `json:"owner"`
	HostPort string  `json:"host_port"`
	Status   string  `json:"status"`
	CPUUsage float64 `json:"cpu_usage"`
	MemUsage float64 `json:"mem_usage"`
}

type Network struct {
	Id       string `json:"id"`
	Subnet   string `json:"subnet"`
	Name     string `json:"name"`
	Internal bool   `json:"internal"`
	Driver   string `json:"driver"`
	Scope    string `json:"scope"`
}
