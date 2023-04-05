package server

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/router"
	log "github.com/Yeuoly/kisara/src/routine/log"
	"github.com/Yeuoly/kisara/src/types"
	uuid "github.com/satori/go.uuid"
)

/*
	this package is used for server to manage the connection among clients
*/

var clientMap sync.Map
var containerMap sync.Map

var waitConnectionChan = make(chan types.RequestConnect, 100)

type ClientItem struct {
	ClientID      string
	Client        *types.Client
	ClientStatus  *types.ClientStatus
	LastHeartBeat time.Time
}

type ContainerItem struct {
	ClientId    string
	ContainerId string
	Container   *types.Container
}

type KisaraOnNodeConnect func(client_id string, client *types.Client)
type KisaraOnNodeDisconnect func(client_id string, client *types.Client)
type KisaraOnNodeHeartBeat func(client_id string, client *types.Client, status *types.ClientStatus)
type KisaraOnNodeLaunchContainer func(client_id string, client *types.Client, container *types.Container)
type KisaraOnNodeStopContainer func(client_id string, client *types.Client, container *types.Container)

var onNodeConnect []KisaraOnNodeConnect
var onNodeDisconnect []KisaraOnNodeDisconnect
var onNodeHeartBeat []KisaraOnNodeHeartBeat
var onNodeLaunchContainer []KisaraOnNodeLaunchContainer
var onNodeStopContainer []KisaraOnNodeStopContainer

func RegisterOnNodeConnect(f KisaraOnNodeConnect) {
	onNodeConnect = append(onNodeConnect, f)
}

func RegisterOnNodeDisconnect(f KisaraOnNodeDisconnect) {
	onNodeDisconnect = append(onNodeDisconnect, f)
}

func RegisterOnNodeHeartBeat(f KisaraOnNodeHeartBeat) {
	onNodeHeartBeat = append(onNodeHeartBeat, f)
}

func RegisterOnNodeLaunchContainer(f KisaraOnNodeLaunchContainer) {
	onNodeLaunchContainer = append(onNodeLaunchContainer, f)
}

func RegisterOnNodeStopContainer(f KisaraOnNodeStopContainer) {
	onNodeStopContainer = append(onNodeStopContainer, f)
}

func UnsetOnNodeConnect() {
	onNodeConnect = []KisaraOnNodeConnect{}
}

func UnsetOnNodeDisconnect() {
	onNodeDisconnect = []KisaraOnNodeDisconnect{}
}

func UnsetOnNodeHeartBeat() {
	onNodeHeartBeat = []KisaraOnNodeHeartBeat{}
}

func UnsetOnNodeLaunchContainer() {
	onNodeLaunchContainer = []KisaraOnNodeLaunchContainer{}
}

func UnsetOnNodeStopContainer() {
	onNodeStopContainer = []KisaraOnNodeStopContainer{}
}

func AddContainer(container_id string, client_id string, container *types.Container) {
	containerMap.Store(container_id, &ContainerItem{
		ClientId:    client_id,
		ContainerId: container_id,
		Container:   container,
	})
	for _, f := range onNodeLaunchContainer {
		client := GetClient(client_id)
		if client != nil {
			f(client_id, client, container)
		}
	}
}

/*
ret:

	*container, client_id, error
*/
func GetContainer(container_id string) (*types.Container, string, error) {
	if container, ok := containerMap.Load(container_id); ok {
		return container.(*ContainerItem).Container, container.(*ContainerItem).ClientId, nil
	}
	return nil, "", errors.New("container not found")
}

func FlushContainer(client_id string) {
	containerMap.Range(func(key, value interface{}) bool {
		if value.(*ContainerItem).ClientId == client_id {
			DeleteContainer(value.(*ContainerItem).ContainerId)
		}
		return true
	})
}

func DeleteContainer(container_id string) {
	containerMap.Delete(container_id)
	for _, f := range onNodeStopContainer {
		container, client_id, err := GetContainer(container_id)
		if err == nil {
			client := GetClient(client_id)
			if client != nil {
				f(client_id, client, container)
			}
		}
	}
}

func (c *ClientItem) GetDemand() (float64, error) {
	if c.ClientStatus == nil {
		return 0, errors.New("client status is not initialized")
	}

	status := *c.ClientStatus
	if status.NetworkUsage >= 0.9 {
		return 1, nil
	}

	if status.CPUUsage >= 0.98 {
		return 1, nil
	}

	if status.MemoryUsage >= 0.98 {
		return 1, nil
	}

	if status.ContainerUsage >= 1 {
		return 1, nil
	}

	return status.CPUUsage*0.7 + status.MemoryUsage*0.2 + status.ContainerUsage*0.1, nil
}

func AddConnectRequest(req types.RequestConnect) {
	waitConnectionChan <- req
}

func GetConnectRequest() types.RequestConnect {
	return <-waitConnectionChan
}

func GetClient(client_id string) *types.Client {
	if client, ok := clientMap.Load(client_id); ok {
		return client.(*ClientItem).Client
	}
	return nil
}

func UpdateHeartBeat(client_id string) error {
	if client, ok := clientMap.Load(client_id); ok {
		client.(*ClientItem).LastHeartBeat = time.Now()
		return nil
	}
	for _, f := range onNodeHeartBeat {
		client := GetClient(client_id)
		if client != nil {
			status, err := GetClientStatus(client_id)
			if err == nil {
				f(client_id, client, &status)
			}
		}
	}
	return errors.New("client not found")
}

func UpdateClientStatus(client_id string, status types.ClientStatus) error {
	//log.Info("[Connection] Received client status update from client [%s] with cpu usage [%f], memory usage [%f], container usage [%f], network usage [%f], containers [%d]", client_id, status.CPUUsage, status.MemoryUsage, status.ContainerUsage, status.NetworkUsage, status.ContainerNum)
	if client, ok := clientMap.Load(client_id); ok {
		status := status
		client.(*ClientItem).ClientStatus = &status
		return nil
	}
	return errors.New("client not found")
}

func UpdateClientContainer(client_id string) error {
	client := GetClient(client_id)
	if client == nil {
		return errors.New("client not found")
	}

	// get all containers
	resp, err := helper.SendGetAndParse[types.KisaraResponseWrap[types.ResponseListContainer]](
		client.GenerateClientURI(router.URI_CLIENT_LIST_CONTAINER),
		helper.HttpPayloadJson(types.RequestListContainer{
			ClientID: client_id,
		}),
		helper.HttpTimeout(2000),
	)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return errors.New(resp.Data.Error)
	}

	// update container list
	for _, container := range resp.Data.Containers {
		AddContainer(container.Id, client_id, &container)
	}

	return nil
}

func Disconnect(client_id string) {
	clientMap.Delete(client_id)
}

func GetClientStatus(client_id string) (types.ClientStatus, error) {
	if client, ok := clientMap.Load(client_id); ok {
		return *client.(*ClientItem).ClientStatus, nil
	}
	return types.ClientStatus{}, errors.New("client not found")
}

func FetchLowestDemandClient() (types.Client, error) {
	var lowestDemandClient types.Client
	var lowestDemand float64 = math.MaxFloat64
	clientMap.Range(func(key, value interface{}) bool {
		client := value.(*ClientItem)
		demand, err := client.GetDemand()
		if err != nil {
			return true
		} else {
			if demand < lowestDemand {
				lowestDemandClient = *client.Client
				lowestDemand = demand
			}
		}
		return true
	})
	if lowestDemand == math.MaxFloat64 {
		return types.Client{}, errors.New("no client found")
	}
	return lowestDemandClient, nil
}

// Server is the main function of the synergy server, it's non-blocking, call it directly without goroutine
func Server(show_log ...bool) {
	if len(show_log) > 0 && show_log[0] {
		log.SetShowLog(show_log[0])
	}
	// add client listener
	log.Info("[Connection] Start listening for new clients")
	go func() {
		for {
			req := GetConnectRequest()
			log.Info("[Connection] New client connected: %s from %s:%d", req.ClientID, req.ClientIp, req.ClientPort)
			// check if the client is already connected
			if _, ok := clientMap.Load(req.ClientID); ok {
				log.Warn("[Connection] Client %s already connected, ignore this connection", req.ClientID)
				req.Callback(types.ResponseConnect{})
				continue
			}
			client_token := uuid.NewV4().String()
			client := &types.Client{
				ClientID:    req.ClientID,
				ClientToken: client_token,
				ClientIp:    req.ClientIp,
				ClientPort:  req.ClientPort,
			}
			clientMap.Store(req.ClientID, &ClientItem{
				ClientID:      req.ClientID,
				Client:        client,
				LastHeartBeat: time.Now(),
			})
			req.Callback(types.ResponseConnect{
				ClientID:    req.ClientID,
				ClientToken: client_token,
			})
			// on client connected
			for _, f := range onNodeConnect {
				f(req.ClientID, client)
			}
			go handleClientConnection(req.ClientID)
		}
	}()
}

func handleClientConnection(client_id string) {
	// update client containers
	if err := UpdateClientContainer(client_id); err != nil {
		log.Warn("[Connection] Failed to update client containers, error: %s", err.Error())
	}
	timer := time.NewTicker(30 * time.Second)
	defer timer.Stop()
	defer log.Info("[Connection] Client %s disconnected", client_id)
	defer func() {
		for _, f := range onNodeDisconnect {
			client := GetClient(client_id)
			if client != nil {
				f(client_id, client)
			}
		}
	}()
	for range timer.C {
		if client, ok := clientMap.Load(client_id); ok {
			if time.Since(client.(*ClientItem).LastHeartBeat) > 90*time.Second {
				clientMap.Delete(client_id)
				return
			} else if time.Since(client.(*ClientItem).LastHeartBeat) > 40*time.Second {
				log.Warn(
					"[Connection] Client %s has not sent heartbeat for %d seconds, server will lower the priority of this client",
					client_id,
					int(time.Since(client.(*ClientItem).LastHeartBeat).Seconds()),
				)
			}
		} else {
			return
		}
	}
}

func GetNodes() []ClientItem {
	var clients []ClientItem
	clientMap.Range(func(key, value interface{}) bool {
		clients = append(clients, *value.(*ClientItem))
		return true
	})
	return clients
}
