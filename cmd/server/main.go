package main

import (
	"fmt"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	return r
}

func setupConfig() {
	helper.InitServerConfig()
}

func main() {
	setupConfig()
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(fmt.Sprintf(":%d", helper.GetConfigInteger("kisaraServer.port")))
}
