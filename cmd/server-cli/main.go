package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	server_api "github.com/Yeuoly/kisara/src/api"
	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/router/server"
	synergy_server "github.com/Yeuoly/kisara/src/routine/synergy/server"
	"github.com/Yeuoly/kisara/src/types"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	gin.SetMode(gin.ReleaseMode)
	logrus.SetLevel(logrus.PanicLevel)
	// start server
	synergy_server.Server()

	// start cli
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("kisara> ")
			text, _ := reader.ReadString('\n')
			// convert CRLF to LF
			text = strings.Replace(text, "\n", "", -1)
			if text == "exit" {
				os.Exit(0)
			} else if text == "help" {
				fmt.Println("help: show this message")
				fmt.Println("exit: exit kisara")
				fmt.Println("list: list all containers")
				fmt.Println("stop [container_id]: stop a container")
				fmt.Println("start [container_id]: start a container")
				fmt.Println("remove [container_id]: remove a container")
				fmt.Println("networks: list all networks")
				fmt.Println("remove-create [subnet] [name]: create a network")
				fmt.Println("remove-network [network_id]: remove a network")
				fmt.Println("images: list all images")
			} else if text == "list" {
				fmt.Println("Start to list all containers")
				resp, err := server_api.ListContainer(types.RequestListContainer{}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					for _, container := range resp.Containers {
						fmt.Println(container)
					}
				}
			}
		}
	}()

	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", helper.GetConfigInteger("kisaraServer.port")))
}
