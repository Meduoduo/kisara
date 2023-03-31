package server

import (
	"github.com/Yeuoly/kisara/src/router"
	"github.com/gin-gonic/gin"

	server_controller "github.com/Yeuoly/kisara/src/controller/server"
)

func Setup(eng *gin.Engine) {
	eng.POST(router.URI_SERVER_CONNECT, server_controller.HandleConnect)
	eng.POST(router.URI_SERVER_DISCONNECT, server_controller.HandleDisconnect)
	eng.POST(router.URI_SERVER_HEARTBEAT, server_controller.HandleHeartBeat)
	eng.POST(router.URI_SERVER_STATUS, server_controller.HandleRecvStatus)
}
