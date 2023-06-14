package types

import (
	"io"
	"mime/multipart"
)

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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// ClientIp
	ClientIp string `json:"client_ip" form:"client_ip" binding:"required"`
	// ClientPort
	ClientPort int `json:"client_port" form:"client_port" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
}

type ResponseDisconnect struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
}

type RequestStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// CPUUsage is the CPU usage of the client
	CPUUsage float64 `json:"cpu_usage" form:"cpu_usage" binding:"required"`
	// MemoryUsage is the memory usage of the client
	MemoryUsage float64 `json:"memory_usage" form:"memory_usage" binding:"required"`
	// DiskUsage is the disk usage of the client
	DiskUsage float64 `json:"disk_usage" form:"disk_usage" binding:"required"`
	// NetworkUsage is the network usage of the client
	NetworkUsage float64 `json:"network_usage" form:"network_usage" binding:"required"`
	// ContainerNum is the number of containers of the client
	ContainerNum int `json:"container_num" form:"container_num" binding:"required"`
	// ContainerUsage is the usage of containers of the client
	ContainerUsage float64 `json:"container_usage" form:"container_usage" binding:"required"`
}

type ResponseStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
}

type RequestHeartBeat struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
}

type ResponseHeartBeat struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Timestamp is the timestamp of the client, used to check the connection, if the timestamp is not updated for a long time, the client should reconnect
	Timestamp int64 `json:"timestamp"`
}

type RequestLaunchContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// func (c *Docker) CreateContainer(image *Image, uid int, port_protocol string, subnet_name string, module string, env_mount ...map[string]string) (*Container, error)
	// Image is the image of the container
	Image string `json:"image" form:"image" binding:"required"`
	// UID is the uid of the container
	UID int `json:"uid" form:"uid" binding:"required"`
	// PortProtocol is the port protocol of the container
	PortProtocol string `json:"port_protocol" form:"port_protocol" binding:"required"`
	// SubnetName is the subnet name of the container
	SubnetName string `json:"subnet_name" form:"subnet_name" binding:"required"`
	// Module is the module of the container
	Module string `json:"module" form:"module" binding:"required"`
	// EnvMount is the env mount of the container
	EnvMount []map[string]string `json:"env_mount" form:"env_mount"`
}

type ResponseLaunchContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// Response Id
	ResponseId string `json:"response_id" form:"response_id" binding:"required"`
}

type RequestCheckLaunchStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// Response Id
	ResponseId string `json:"response_id" form:"response_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// ContainerID is the container ID of the container
	ContainerID string `json:"container_id" form:"container_id" binding:"required"`
}

type ResponseStopContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestRemoveContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// ContainerID is the container ID of the container
	ContainerID string `json:"container_id" form:"container_id" binding:"required"`
}

type ResponseRemoveContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestListContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// func (c *Docker) Exec(container_id string, cmd string) error
	// ContainerID is the container ID of the container
	ContainerID string `json:"container_id" form:"container_id" binding:"required"`
	// Cmd is the cmd of the container
	Cmd string `json:"cmd" form:"cmd" binding:"required"`
}

type ResponseExecContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestInspectContainer struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// func (c *Docker) InspectContainer(container_id string, has_state ...bool) (*kisara_types.Container, error)
	// ContainerID is the container ID of the container
	ContainerIDs []string `json:"container_id" form:"container_id" binding:"required"`
	// HasState is the has state of the container
	HasState bool `json:"has_state" form:"has_state" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// func (c *Docker) DeleteImage(image_id string) error
	// ImageID is the image ID of the container
	ImageID string `json:"image_id" form:"image_id" binding:"required"`
}

type ResponseDeleteImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestPullImage struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// func HandleControllerRequestPullImage(request_id string, image_name string, port_protocol string, user string)
	// ImageName is the image name of the container
	ImageName string `json:"image_name" form:"image_name" binding:"required"`
	// PortProtocol is the port protocol of the container
	PortProtocol string `json:"port_protocol" form:"port_protocol" binding:"required"`
	// User is the user of the container
	User string `json:"user" form:"user" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// func HandleControllerRequestCheckPullImage(request_id string)
	// RequestID is the request ID of the container
	MessageResponseId string `json:"message_resposne_id" form:"message_resposne_id" binding:"required"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id" form:"finish_response_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// func (c *Docker) CreateNetwork(subnet string, name string, host_join bool) error
	// Subnet is the subnet of the container
	Subnet string `json:"subnet" form:"subnet" binding:"required"`
	// Name is the name of the container
	Name string `json:"name" form:"name" binding:"required"`
	// Internal is the host join of the container
	Internal bool `json:"internal" form:"internal" binding:"required"`
	// Driver
	Driver string `json:"driver" form:"driver" binding:"required"`
}

type ResponseCreateNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestRemoveNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// func (c *Docker) RemoveNetwork(network_id string) error
	// NetworkID is the network ID of the container
	NetworkID string `json:"network_id" form:"network_id" binding:"required"`
}

type ResponseRemoveNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the container
	Error string `json:"error"`
}

type RequestListNetwork struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// ServiceConfig
	ServiceConfig KisaraService `json:"service_config" form:"service_config" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// RequestID is the request ID of the container
	MessageResponseId string `json:"response_id" form:"response_id" binding:"required"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id" form:"finish_response_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// ServiceID
	ServiceID string `json:"service_id" form:"service_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// ResponseID
	ResponseID string `json:"response_id" form:"response_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// Vm id
	ImageId string `json:"image_id" form:"image_id" binding:"required"`
	// Cpu limit
	CpuLimit int `json:"cpu_limit" form:"cpu_limit" binding:"required"`
	// Memory limit
	MemoryLimit int `json:"memory_limit" form:"memory_limit" binding:"required"`
	// Disk limit
	DiskLimit int `json:"disk_limit" form:"disk_limit" binding:"required"`
	// Network limit
	NetworkLimit int `json:"network_limit" form:"network_limit" binding:"required"`
	// NetworkNames is the network names of the container, vm will be connected to these networks if they exist, otherwise, error will be returned
	NetworkNames []string `json:"network_names" form:"network_names" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// RequestID is the request ID of the container
	MessageResponseId string `json:"response_id" form:"response_id" binding:"required"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id" form:"finish_response_id" binding:"required"`
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
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
}

type ResponseListVm struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Vms is the vms of the container
	Vms []VM `json:"vms"`
	// Error is the error of the container
	Error string `json:"error"`
}

// send request
type RequestNetworkMonitorRun struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// context
	Context io.Reader `json:"context" form:"context"`
	// NetworkName
	NetworkName string `json:"network_name" form:"network_name" binding:"required"`
}

// parsed request
type RequestNetworkMonitorRunToByRecieved struct {
	// ClientID is the unique ID of the client
	ClientID string `form:"client_id" binding:"required"`
	// context
	Context *multipart.FileHeader `form:"file" binding:"required"`
	// NetworkName
	NetworkName string `form:"network_name" binding:"required"`
}

type ResponseNetworkMonitorRun struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the vm
	Error string `json:"error"`
	// Response Id
	ResponseId string `json:"response_id"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id"`
}

type RequestNetworkMonitorCheck struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// Response Id
	ResponseId string `json:"response_id" form:"response_id" binding:"required"`
	// FinishResponseID is the finish response ID of the container
	FinishResponseID string `json:"finish_response_id" form:"finish_response_id" binding:"required"`
}

type ResponseNetworkMonitorCheck struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the run
	Error string `json:"error"`
	// Finished is the finished of the run
	Finished bool `json:"finished"`
	// Message
	Message string `json:"message"`
	// NetworkMonitor Container Id
	NetworkMonitorContainerId string `json:"network_monitor_container_id"`
}

type ResponseFinalNetworkMonitorStatus struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the run
	Error string `json:"error"`
	// NetworkMonitor Container Id
	NetworkMonitorContainerId string `json:"network_monitor_container_id"`
}

type RequestNetworkMonitorStop struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// NetworkMonitor Container Id
	NetworkMonitorContainerId string `json:"network_monitor_container_id" form:"network_monitor_container_id" binding:"required"`
}

type ResponseNetworkMonitorStop struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the run
	Error string `json:"error"`
}

type RequestNetworkMonitorRunScript struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id" form:"client_id" binding:"required"`
	// Containers to be tested
	Containers KisaraNetworkTestSet `json:"containers" form:"containers" binding:"required"`
}

type ResponseNetworkMonitorRunScript struct {
	// ClientID is the unique ID of the client
	ClientID string `json:"client_id"`
	// Error is the error of the run
	Error string `json:"error"`
	// Containers to be tested
	Result KisaraNetworkTestResultSet `json:"result"`
}
