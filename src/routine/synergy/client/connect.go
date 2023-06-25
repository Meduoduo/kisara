package client

/*
	this package is used for client to manage the connection with the server
*/

import (
	"fmt"
	"math"
	"time"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/router"
	docker "github.com/Yeuoly/kisara/src/routine/docker"
	log "github.com/Yeuoly/kisara/src/routine/log"
	"github.com/Yeuoly/kisara/src/types"
	uuid "github.com/satori/go.uuid"
)

var clientId string
var clientIp string
var clientPort int
var clientToken string
var serverIp string
var serverPort int

func GetClientId() string {
	return clientId
}

func getServerRequest(uri string) string {
	return fmt.Sprintf("http://%s:%d%s", serverIp, serverPort, uri)
}

// upload client status to server
var (
	cpu_usage_arr       []float64
	cpu_usage_arr_index int
)

const (
	CPU_USAGE_ARR_SIZE = 10
)

func setCPUUsageArr(cpu_usage float64) {
	cpu_usage_arr_index = (cpu_usage_arr_index + 1) % CPU_USAGE_ARR_SIZE
	cpu_usage_arr[cpu_usage_arr_index] = cpu_usage
}

func getCPUUsage() float64 {
	cpu_usage, err := helper.GetCPUUsageTotal()
	if err != nil {
		log.Warn("[Connection] Failed to get CPU usage: %s", err.Error())
		cpu_usage = 0
	}

	setCPUUsageArr(cpu_usage)

	sum := 0.0
	for _, v := range cpu_usage_arr {
		sum += v
	}
	return sum / CPU_USAGE_ARR_SIZE
}

// Client is the main function of the synergy client, it's non-blocking, call it directly without goroutine
func Client() {
	log.Info("[Connection] Start continous connection with server")
	log.Info("[Connection] Initializing client")
	clientIp = helper.GetConfigString("kisaraClient.address")
	clientPort = helper.GetConfigInteger("kisaraClient.port")
	if clientIp == "" {
		log.Panic("[Connection] Client IP is not set")
	} else if clientPort == 0 {
		log.Panic("[Connection] Client port is not set")
	}

	serverIp = helper.GetConfigString("kisaraServer.address")
	serverPort = helper.GetConfigInteger("kisaraServer.port")
	if serverIp == "" {
		log.Panic("[Connection] Server IP is not set")
	} else if serverPort == 0 {
		log.Panic("[Connection] Server port is not set")
	}

	// init cpu usage array
	cpu_usage, err := helper.GetCPUUsageTotal()
	if err != nil {
		log.Warn("[Connection] Failed to get CPU usage: %s", err.Error())
		cpu_usage = 0
	}
	cpu_usage_arr = make([]float64, CPU_USAGE_ARR_SIZE)
	for i := 0; i < CPU_USAGE_ARR_SIZE; i++ {
		setCPUUsageArr(cpu_usage)
	}

	clientId = uuid.NewV4().String()

	log.Info("[Connection] Finished Initialize client with client id : %s", clientId)
	log.Info("[Connection] Make sure use client id %s in server, or the connection may be considered as a unAuthorized connection", clientId)
	go func() {
		for {
			log.Info("[Connection] Connecting to server %s:%d", serverIp, serverPort)
			connect()
			time.Sleep(time.Duration(5 * time.Second))
		}
	}()
}

func uploadStatus() {
	cpu_usage := getCPUUsage()

	mem_usage, _, _, err := helper.GetMemUsage()
	if err != nil {
		log.Warn("[Connection] Failed to get memory usage: %s", err.Error())
		mem_usage = 0
	}

	disk_usage, _, _, err := helper.GetDiskUsage()
	if err != nil {
		log.Warn("[Connection] Failed to get disk usage: %s", err.Error())
		disk_usage = 0
	}

	network_usage_in, _, err := helper.GetNetUsagePercent()
	if err != nil {
		log.Warn("[Connection] Failed to get network usage: %s", err.Error())
		network_usage_in = 0
	}

	max_container := helper.GetMaxContainer()
	if max_container == 0 {
		log.Warn("[Connection] Max container is not set")
		max_container = 1
	}

	docker := docker.NewDocker()
	container_num, err := docker.GetContainerNumber()
	if err != nil {
		log.Warn("[Connection] Failed to get container number: %s", err.Error())
		container_num = 0
	}

	//log.Info("[Connection] Uploading status to server %s:%d with cpu %f%%, mem %f%%, disk %f%%, net %f%%, container_num %d", serverIp, serverPort, cpu_usage, mem_usage, disk_usage, network_usage_in, container_num)

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseStatus]](
		getServerRequest(router.URI_SERVER_STATUS),
		helper.HttpPayloadJson(types.RequestStatus{
			ClientID:       clientId,
			CPUUsage:       math.Round(cpu_usage*100) / 100,
			MemoryUsage:    math.Round(mem_usage*100) / 100,
			DiskUsage:      math.Round(disk_usage*100) / 100,
			NetworkUsage:   math.Round(network_usage_in*100) / 100,
			ContainerNum:   container_num,
			ContainerUsage: math.Round(float64(container_num)/float64(max_container)*100) / 100,
		}),
		helper.HttpTimeout(5000),
	)

	if err != nil {
		log.Warn("[Connection] Failed to upload status to server: %s", err.Error())
		return
	}
	if resp.Code != 0 {
		log.Warn("[Connection] Failed to upload status to server: %s", resp.Message)
		return
	}
}

func connect() {
	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseConnect]](
		getServerRequest(router.URI_SERVER_CONNECT),
		helper.HttpPayloadJson(types.RequestConnect{
			ClientID:   clientId,
			ClientIp:   clientIp,
			ClientPort: clientPort,
		}),
		helper.HttpTimeout(5000),
	)
	if err != nil {
		log.Error("[Connection] Failed to connect to server: %s", err.Error())
		return
	}
	if resp.Code != 0 {
		log.Error("[Connection] Failed to connect to server: %s", resp.Message)
		return
	}
	clientToken = resp.Data.ClientToken
	if clientToken == "" {
		log.Error("[Connection] Failed to connect to server: token is empty")
		return
	}
	if resp.Data.ClientID != clientId {
		log.Error("[Connection] Failed to connect to server: client id is not matched")
		return
	}
	// start heart beat
	log.Info("[Connection] Connected to server, start heart beat")
	defer log.Warn("[Connection] Heart beat stopped")
	// start status monitor
	ticker := time.NewTicker(time.Duration(10 * time.Second))
	defer ticker.Stop()
	uploadStatus()

	go func() {
		for range ticker.C {
			uploadStatus()
		}
	}()

	for {
		heart_beat := func() bool {
			// send heart beat at most 3 times, if failed, reconnect
			for i := 0; i < 3; i++ {
				resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseHeartBeat]](
					getServerRequest(router.URI_SERVER_HEARTBEAT),
					helper.HttpPayloadJson(types.RequestHeartBeat{
						ClientID: clientId,
					}),
					helper.HttpTimeout(5000),
				)
				if err != nil {
					log.Error("[Connection] Failed to send heart beat: %s", err.Error())
					continue
				}
				if resp.Code != 0 {
					log.Error("[Connection] Failed to send heart beat: %s", resp.Message)
					continue
				}
				if resp.Data.ClientID != clientId {
					log.Error("[Connection] Failed to send heart beat: client id is not matched")
					continue
				}
				current_timestamp := time.Now().Unix()
				if math.Abs(float64(current_timestamp-resp.Data.Timestamp)) > 90 {
					log.Error("[Connection] Failed to send heart beat: server may be down")
					continue
				}
				log.Info("[Connection] Heart beat……")
				time.Sleep(30 * time.Second)

				return true
			}
			return false
		}

		if !heart_beat() {
			log.Error("[Connection] Heart beat failed, reconnecting")
			return
		}
	}
}
