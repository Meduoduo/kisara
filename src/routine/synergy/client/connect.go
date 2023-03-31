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

func getServerRequest(uri string) string {
	return fmt.Sprintf("http://%s:%d%s", serverIp, serverPort, uri)
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
	for {
		resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseHeartBeat]](
			getServerRequest(router.URI_SERVER_HEARTBEAT),
			helper.HttpPayloadJson(types.RequestHeartBeat{
				ClientID: clientId,
			}),
			helper.HttpTimeout(5000),
		)
		if err != nil {
			log.Error("[Connection] Failed to send heart beat: %s", err.Error())
			return
		}
		if resp.Code != 0 {
			log.Error("[Connection] Failed to send heart beat: %s", resp.Message)
			return
		}
		if resp.Data.ClientID != clientId {
			log.Error("[Connection] Failed to send heart beat: client id is not matched")
			return
		}
		current_timestamp := time.Now().Unix()
		if math.Abs(float64(current_timestamp-resp.Data.Timestamp)) > 90 {
			log.Error("[Connection] Failed to send heart beat: server may be down")
			return
		}
		log.Info("[Connection] Heart beat……")
		time.Sleep(30 * time.Second)
	}
}
