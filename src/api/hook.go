package api

import (
	"github.com/Yeuoly/kisara/src/routine/synergy/server"
)

func RegisterOnNodeConnect(f server.KisaraOnNodeConnect) {
	server.RegisterOnNodeConnect(f)
}

func RegisterOnNodeDisconnect(f server.KisaraOnNodeDisconnect) {
	server.RegisterOnNodeDisconnect(f)
}

func RegisterOnNodeHeartBeat(f server.KisaraOnNodeHeartBeat) {
	server.RegisterOnNodeHeartBeat(f)
}

func RegisterOnNodeLaunchContainer(f server.KisaraOnNodeLaunchContainer) {
	server.RegisterOnNodeLaunchContainer(f)
}

func RegisterOnNodeStopContainer(f server.KisaraOnNodeStopContainer) {
	server.RegisterOnNodeStopContainer(f)
}

func UnsetOnNodeConnect() {
	server.UnsetOnNodeConnect()
}

func UnsetOnNodeDisconnect() {
	server.UnsetOnNodeDisconnect()
}

func UnsetOnNodeHeartBeat() {
	server.UnsetOnNodeHeartBeat()
}

func UnsetOnNodeLaunchContainer() {
	server.UnsetOnNodeLaunchContainer()
}

func UnsetOnNodeStopContainer() {
	server.UnsetOnNodeStopContainer()
}
