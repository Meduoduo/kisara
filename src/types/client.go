package types

type Client struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// ClientIp
	ClientIp string `json:"client_ip"`
	// ClientPort
	ClientPort int `json:"client_port"`
	// Token
	ClientToken string `json:"client_token"`
}

type ClientStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// CPUUsage is the CPU usage of the client
	CPUUsage float64 `json:"cpu_usage"`
	// MemoryUsage is the memory usage of the client
	MemoryUsage float64 `json:"memory_usage"`
	// DiskUsage is the disk usage of the client
	DiskUsage float64 `json:"disk_usage"`
	// NetworkUsage is the network usage of the client
	NetworkUsage float64 `json:"network_usage"`
	// ContainerNum is the number of containers of the client
	ContainerNum int `json:"container_num"`
}
