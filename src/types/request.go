package types

type KisaraReponse struct {
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

func SuccessResponse(data interface{}) KisaraReponse {
	return KisaraReponse{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func ErrorResponse(code int, message string) KisaraReponse {
	if code >= 0 {
		code = -1
	}
	return KisaraReponse{
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
