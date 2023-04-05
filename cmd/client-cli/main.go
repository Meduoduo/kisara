package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
	gin.SetMode(gin.ReleaseMode)
	// start client
	synergy_client.Client()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("kisara> ")
			text, _ := reader.ReadString('\n')
			// convert CRLF to LF
			text = strings.Replace(text, "\n", "", -1)
			if text == "exit" {
				os.Exit(0)
			}
			fmt.Println(text)
		}
	}()

	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", helper.GetConfigInteger("kisaraClient.port")))
}
