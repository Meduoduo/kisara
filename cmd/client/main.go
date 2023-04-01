package main

import (
	"fmt"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/router/client"
	synergy_client "github.com/Yeuoly/kisara/src/routine/synergy/client"
	takina "github.com/Yeuoly/kisara/src/routine/takina"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	client.Setup(r)

	return r
}

func setupConfig() {
	helper.InitServerConfig()
	takina.InitTakina()
}

func main() {
	setupConfig()
	if helper.GetConfigString("kisara.mode") == "dev" {
		gin.SetMode(gin.DebugMode)
	} else if helper.GetConfigString("kisara.mode") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	// start client
	synergy_client.Client()
	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", helper.GetConfigInteger("kisaraClient.port")))
}
