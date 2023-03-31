package server

import (
	"sync"
	"time"

	log "github.com/Yeuoly/kisara/src/routine/log"
	"github.com/Yeuoly/kisara/src/types"
	uuid "github.com/satori/go.uuid"
)

/*
	this package is used for server to manage the connection among clients
*/

var clientMap sync.Map

var waitConnectionChan = make(chan types.RequestConnect, 100)

type ClientItem struct {
	ClientID      string
	Client        *types.Client
	LastHeartBeat time.Time
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

func UpdateHeartBeat(client_id string) {
	if client, ok := clientMap.Load(client_id); ok {
		client.(*ClientItem).LastHeartBeat = time.Now()
	}
}

func Disconnect(client_id string) {
	clientMap.Delete(client_id)
}

// Server is the main function of the synergy server, it's non-blocking, call it directly without goroutine
func Server() {
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
			go handleClientConnection(req.ClientID)
		}
	}()
}

func handleClientConnection(client_id string) {
	timer := time.NewTicker(30 * time.Second)
	defer timer.Stop()
	defer log.Info("[Connection] Client %s disconnected", client_id)
	for range timer.C {
		if client, ok := clientMap.Load(client_id); ok {
			if time.Since(client.(*ClientItem).LastHeartBeat) > 90*time.Second {
				clientMap.Delete(client_id)
				return
			} else if time.Since(client.(*ClientItem).LastHeartBeat) > 40*time.Second {
				log.Warn("[Connection] Client %s has not sent heartbeat for 40 seconds, server will lower the priority of this client", client_id)
				return
			}
		} else {
			return
		}
	}
}
