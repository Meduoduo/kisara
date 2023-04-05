package api

import (
	"fmt"
	"os"

	"github.com/Yeuoly/kisara/src/helper"
	server_routes "github.com/Yeuoly/kisara/src/router/server"
	log "github.com/Yeuoly/kisara/src/routine/log"
	"github.com/Yeuoly/kisara/src/routine/synergy/server"
	"github.com/gin-gonic/gin"
)

// LaunchKisaraServer launches the Kisara server, it's non-blocking
func LaunchKisaraServer(ignoreLogInfo bool) {
	go launchKisaraServerImpl(ignoreLogInfo)
}

func setupConfig() {
	helper.InitServerConfig()
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	server_routes.Setup(r)
	return r
}

func launchKisaraServerImpl(ignoreLogInfo bool) {
	var err error

	if ignoreLogInfo {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
		gin.DefaultWriter, err = os.Create(os.DevNull)
		if err != nil {
			log.Panic("[Kisara] Failed to ignore log info: " + err.Error())
		}
	}

	setupConfig()

	server.Server()
	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", helper.GetConfigInteger("kisaraServer.port")))
}
