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
	gin.DisableConsoleColor()
	gin.DefaultWriter, _ = os.Create(os.DevNull)

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
				fmt.Println("start [image_name]: start a container")
				fmt.Println("remove [container_id]: remove a container")
				fmt.Println("exec [container_id] cmd: execute command in container")
				fmt.Println("inspect [container_id]: inspect a container")
				fmt.Println("networks: list all networks")
				fmt.Println("create-network [client_id] [subnet] [name]: create a network")
				fmt.Println("remove-network [client_id] [network_id]: remove a network")
				fmt.Println("images: list all images")
				fmt.Println("pull-image [client_id] [image_name]: pull an image")
				fmt.Println("delete-image [client_id] [image_id]: remove an image")
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
			} else if strings.HasPrefix(text, "stop ") {
				fmt.Println("Start to stop container")
				resp, err := server_api.StopContainer(types.RequestStopContainer{
					ContainerID: strings.TrimPrefix(text, "stop "),
				}, time.Duration(time.Second*30))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println(resp)
				}
			} else if strings.HasPrefix(text, "start ") {
				fmt.Println("Start to start container")
				resp, err := server_api.LaunchContainer(types.RequestLaunchContainer{
					Image:        strings.TrimPrefix(text, "start "),
					UID:          9,
					PortProtocol: "80/tcp",
					SubnetName:   "irina-train",
					Module:       "train",
				}, time.Duration(time.Second*30))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println(resp)
				}
			} else if strings.HasPrefix(text, "remove ") {
				fmt.Println("Start to remove container")
				resp, err := server_api.RemoveContainer(types.RequestRemoveContainer{
					ContainerID: strings.TrimPrefix(text, "remove "),
				}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println(resp)
				}
			} else if strings.HasPrefix(text, "exec") {
				fmt.Println("Start to exec command in container")
				split := strings.Split(strings.TrimPrefix(text, "exec "), " ")
				if len(split) < 2 {
					fmt.Println("Error: invalid command")
					continue
				}
				resp, err := server_api.ExecContainer(types.RequestExecContainer{
					ContainerID: split[0],
					Cmd:         strings.Join(split[1:], " "),
				}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println(resp)
				}
			} else if strings.HasPrefix(text, "inspect") {
				fmt.Println("Start to inspect container")
				resp, err := server_api.InspectContainer(types.RequestInspectContainer{
					ContainerIDs: []string{strings.TrimPrefix(text, "inspect ")},
				}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println(resp)
				}
			} else if text == "images" {
				fmt.Println("Start to list all images")
				resp, err := server_api.ListImage(types.RequestListImage{}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					for _, image := range resp.Images {
						fmt.Println(image)
					}
				}
			} else if text == "networks" {
				fmt.Println("Start to list all networks")
				resp, err := server_api.ListNetwork(types.RequestListNetwork{}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					for _, network := range resp.Networks {
						fmt.Println(network)
					}
				}
			} else if strings.HasPrefix(text, "create-network ") {
				fmt.Println("Start to create network")
				split := strings.Split(strings.TrimPrefix(text, "create-network "), " ")
				if len(split) != 3 {
					fmt.Println("Error: invalid command")
					continue
				}
				resp, err := server_api.CreateNetwork(types.RequestCreateNetwork{
					ClientID: split[0],
					Subnet:   split[1],
					Name:     split[2],
				}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println(resp)
				}
			} else if strings.HasPrefix(text, "remove-network ") {
				fmt.Println("Start to remove network")
				split := strings.Split(strings.TrimPrefix(text, "remove-network "), " ")
				if len(split) != 2 {
					fmt.Println("Error: invalid command")
					continue
				}
				resp, err := server_api.RemoveNetwork(types.RequestRemoveNetwork{
					ClientID:  split[0],
					NetworkID: split[1],
				}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println(resp)
				}
			} else if text == "images" {
				fmt.Println("Start to list all images")
				resp, err := server_api.ListImage(types.RequestListImage{}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					for _, image := range resp.Images {
						fmt.Println(image)
					}
				}
			} else if strings.HasPrefix(text, "pull-image") {
				fmt.Println("Start to pull image")
				split := strings.Split(strings.TrimPrefix(text, "pull-image "), " ")
				if len(split) != 2 {
					fmt.Println("Error: invalid command")
					continue
				}
				resp, err := server_api.PullImage(types.RequestPullImage{
					ClientID:  split[0],
					ImageName: split[1],
				}, time.Duration(time.Second*600), func(message string) {
					fmt.Printf("Pulling image: %s\r\n", message)
				})
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println(resp)
				}
			} else if strings.HasPrefix(text, "delete-image") {
				fmt.Println("Start to delete image")
				split := strings.Split(strings.TrimPrefix(text, "delete-image "), " ")
				if len(split) != 2 {
					fmt.Println("Error: invalid command")
					continue
				}
				resp, err := server_api.DeleteImage(types.RequestDeleteImage{
					ClientID: split[0],
					ImageID:  split[1],
				}, time.Duration(time.Second*5))
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println(resp)
				}
			} else if text == "nodes" {
				fmt.Println("Start to list all nodes")
				resp, err := server_api.GetNodes()
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					for _, node := range resp {
						fmt.Println(node)
					}
				}
			} else {
				fmt.Println("Unknown command, type \"help\" to show help message")
			}
		}
	}()

	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", helper.GetConfigInteger("kisaraServer.port")))
}
