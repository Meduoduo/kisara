package client

import (
	"fmt"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/router/client"
	synergy_client "github.com/Yeuoly/kisara/src/routine/synergy/client"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	client.Setup(r)
	return r
}

func Main() {
	setup()
	attachTakinaHook()
	initDocker()
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
