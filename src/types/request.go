package types

type KisaraResponse struct {
	// Code is the code of the response
	Code int `json:"code"`
	// Message is the message of the response
	Message string `json:"message"`
	// Data is the data of the response
	Data interface{} `json:"data"`
}

type KisaraResponseWrap[T any] struct {
	// Code is the code of the response
	Code int `json:"code"`
	// Message is the message of the response
	Message string `json:"message"`
	// Data is the data of the response
	Data T `json:"data"`
}

func SuccessResponse(data interface{}) KisaraResponse {
	return KisaraResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func ErrorResponse(code int, message string) KisaraResponse {
	if code >= 0 {
		code = -1
	}
	return KisaraResponse{
		Code:    code,
		Message: message,
	}
}

type RequestConnect struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// ClientIp
	ClientIp string `json:"client_ip"`
	// ClientPort
	ClientPort int `json:"client_port"`
	// callback
	Callback func(ResponseConnect) `json:"-"`
}

type ResponseConnect struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Token is the token of the client
	ClientToken string `json:"client_token"`
}

type RequestDisconnect struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
}

type ResponseDisconnect struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
}

type RequestStatus struct {
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
	// ContainerUsage is the usage of containers of the client
	ContainerUsage float64 `json:"container_usage"`
}

type ResponseStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
}

type RequestHeartBeat struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
}

type ResponseHeartBeat struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Timestamp is the timestamp of the client, used to check the connection, if the timestamp is not updated for a long time, the client should reconnect
	Timestamp int64 `json:"timestamp"`
}

type RequestLaunchContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func (c *Docker) CreateContainer(image *Image, uid int, port_protocol string, subnet_name string, module string, env_mount ...map[string]string) (*Container, error)
	// Image is the image of the container
	Image string `json:"image"`
	// UID is the uid of the container
	UID int `json:"uid"`
	// PortProtocol is the port protocol of the container
	PortProtocol string `json:"port_protocol"`
	// SubnetName is the subnet name of the container
	SubnetName string `json:"subnet_name"`
	// Module is the module of the container
	Module string `json:"module"`
	// EnvMount is the env mount of the container
	EnvMount []map[string]string `json:"env_mount"`
}

type ResponseLaunchContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Response Id
	ResponseId string `json:"response_id"`
}

type RequestCheckLaunchStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Response Id
	ResponseId string `json:"response_id"`
}

type ResponseCheckLaunchStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// ContainerID is the container ID of the container
	Container Container
	// Error is the error of the container
	Error string `json:"error"`
	// Finished
	Finished bool `json:"finished"`
}

type ResponseFinalLaunchStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// ContainerID is the container ID of the container
	Container Container
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestStopContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// ContainerID is the container ID of the container
	ContainerID string `json:"container_id"`
}

type ResponseStopContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestRemoveContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// ContainerID is the container ID of the container
	ContainerID string `json:"container_id"`
}

type ResponseRemoveContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestListContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func (c *Docker) ListContainer() (*[]*kisara_types.Container, error)
}

type ResponseListContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Containers is the containers of the container
	Containers []Container `json:"containers"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestExecContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func (c *Docker) Exec(container_id string, cmd string) error
	// ContainerID is the container ID of the container
	ContainerID string `json:"container_id"`
	// Cmd is the cmd of the container
	Cmd string `json:"cmd"`
}

type ResponseExecContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestInspectContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func (c *Docker) InspectContainer(container_id string, has_state ...bool) (*kisara_types.Container, error)
	// ContainerID is the container ID of the container
	ContainerIDs []string `json:"container_id"`
	// HasState is the has state of the container
	HasState bool `json:"has_state"`
}

type ResponseInspectContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Container is the container of the container
	Containers []Container `json:"container"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestListImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func (c *Docker) ListImage() (*[]*kisara_types.Image, error)
}

type ResponseListImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Images is the images of the container
	Images []Image `json:"images"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestDeleteImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func (c *Docker) DeleteImage(image_id string) error
	// ImageID is the image ID of the container
	ImageID string `json:"image_id"`
}

type ResponseDeleteImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestPullImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func HandleControllerRequestPullImage(request_id string, image_name string, port_protocol string, user string)
	// ImageName is the image name of the container
	ImageName string `json:"image_name"`
	// PortProtocol is the port protocol of the container
	PortProtocol string `json:"port_protocol"`
	// User is the user of the container
	User string `json:"user"`
}

type ResponsePullImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
	// RequestID is the request ID of the container
	MessageResponseId string `json:"message_resposne_id"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id"`
}

type RequestCheckPullImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func HandleControllerRequestCheckPullImage(request_id string)
	// RequestID is the request ID of the container
	MessageResponseId string `json:"message_resposne_id"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id"`
}

type ResponseCheckPullImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
	// Finished is the finished of the container
	Finished bool `json:"finished"`
	// RequestID is the request ID of the container
	Message string `json:"message"`
}

type ResponseFinalPullImageStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestCreateNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func (c *Docker) CreateNetwork(subnet string, name string, host_join bool) error
	// Subnet is the subnet of the container
	Subnet string `json:"subnet"`
	// Name is the name of the container
	Name string `json:"name"`
	// Internal is the host join of the container
	Internal bool `json:"internal"`
	// Driver
	Driver string `json:"driver"`
}

type ResponseCreateNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestRemoveNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func (c *Docker) RemoveNetwork(network_id string) error
	// NetworkID is the network ID of the container
	NetworkID string `json:"network_id"`
}

type ResponseRemoveNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestListNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// func (c *Docker) ListNetwork() (*[]*types.Network, error)
}

type ResponseListNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Networks is the networks of the container
	Networks []Network `json:"networks"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestLaunchService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// ServiceConfig
	ServiceConfig KisaraService `json:"service_config"`
}

type ResponseLaunchService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
	// Service
	MessageResponseId string `json:"response_id"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id"`
}

type RequestCheckLaunchService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// RequestID is the request ID of the container
	MessageResponseId string `json:"response_id"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id"`
}

type ResponseCheckLaunchService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
	// Finished is the finished of the container
	Finished bool `json:"finished"`
	// Message
	Message string `json:"message"`
	// Service
	Service Service `json:"service"`
}

type ResponseFinalLaunchServiceStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
	// Service
	Service Service `json:"service"`
}

type RequestListService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
}

type ResponseListService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Services is the services of the container
	Services []Service `json:"services"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestStopService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// ServiceID
	ServiceID string `json:"service_id"`
}

type ResponseStopService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
	// ResponseID
	ResponseID string `json:"response_id"`
}

type RequestCheckStopService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// ResponseID
	ResponseID string `json:"response_id"`
}

type ResponseCheckStopService struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
	// Finished is the finished of the container
	Finished bool `json:"finished"`
}

// request to launch a vm, with the vm image id, cpu limit, memory limit, disk limit, network limit
// if limit is 0, it's unlimited
type RequestLaunchVm struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Vm id
	ImageId string `json:"image_id"`
	// Cpu limit
	CpuLimit int `json:"cpu_limit"` // 0.5 = 50% of 1 core
	// Memory limit
	MemoryLimit int `json:"memory_limit"` // bytes
	// Disk limit
	DiskLimit int `json:"disk_limit"` // bytes
	// Network limit
	NetworkLimit int `json:"network_limit"` // bytes per second
	// NetworkNames is the network names of the container, vm will be connected to these networks if they exist, otherwise, error will be returned
	NetworkNames []string `json:"network_names"`
}

type ResponseLaunchVm struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the vm
	Error string `json:"error"`
	// Vm
	MessageResponseId string `json:"response_id"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id"`
}

type RequestCheckLaunchVm struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// RequestID is the request ID of the container
	MessageResponseId string `json:"response_id"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id"`
}

type ResponseCheckLaunchVm struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the vm
	Error string `json:"error"`
	// Finished is the finished of the vm
	Finished bool `json:"finished"`
	// Message
	Message string `json:"message"`
}

type ResponseFinalLaunchVmStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the vm
	Error string `json:"error"`
	// Vm
	Vm VM `json:"vm"`
}

type RequestListVm struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
}

type ResponseListVm struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Vms is the vms of the container
	Vms []VM `json:"vms"`
	// Error is the error of the container
	Error string `json:"error"`
}
