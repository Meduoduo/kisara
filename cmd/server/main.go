package main

import (
	"fmt"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/router/server"
	synergy_server "github.com/Yeuoly/kisara/src/routine/synergy/server"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	server.Setup(r)

	return r
}

func setupConfig() {
	helper.InitServerConfig()
}

func main() {
	setupConfig()
	if helper.GetConfigString("kisara.mode") == "dev" {
		gin.SetMode(gin.DebugMode)
	} else if helper.GetConfigString("kisara.mode") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	// start server
	synergy_server.Server()
	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", helper.GetConfigInteger("kisaraServer.port")))
}
