package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Yeuoly/kisara/src/client"
	"github.com/Yeuoly/kisara/src/routine/docker"
	"github.com/Yeuoly/kisara/src/types"
	uuid "github.com/satori/go.uuid"
)

func main() {
	go client.Main()
	time.Sleep(time.Second * 20)
	testService()

	select {}
}

func testService() {
	service_config := types.ServiceConfig{
		Containers: []types.ServiceConfigContainer{
			{
				Image: "yeuoly/awd_training_ping:v1",
				Ports: []types.ServiceConfigContainerPortMapping{
					{
						Port:     22,
						Protocol: "tcp",
					},
				},
				Networks: []types.ServiceConfigContainerNetwork{
					{
						Network:    "A",
						RandomCIDR: true,
					},
				},
				Env: map[string]string{},
			},
			{
				Image: "yeuoly/awd_training_ping:v1",
				Ports: []types.ServiceConfigContainerPortMapping{},
				Networks: []types.ServiceConfigContainerNetwork{
					{
						Network:    "A",
						RandomCIDR: true,
					},
				},
				Flags: []types.ServiceConfigContainerFlag{
					{
						FlagCommand: "echo $flag > /flag",
						FlagScore:   100,
						FlagUuid:    uuid.NewV4().String(),
					},
				},
				Env: map[string]string{},
			},
		},
		TotalScore:     100,
		NetworkCount:   1,
		ContainerCount: 2,
	}

	service_json, _ := json.Marshal(service_config)
	service := types.KisaraService{
		Id:          uuid.NewV4().String(),
		Name:        "service-test",
		Description: "service-test",
		Owner:       9,
		Config:      string(service_json),
	}

	client := docker.NewDocker()
	defer client.Stop()

	images, _ := client.ListImage()
	for _, image := range *images {
		if image.Name == "yeuoly/awd_training_ping:v1" {
			client.DeleteImage(image.Uuid)
		}
	}

	service_resp, err := client.CreateService(service)
	if err != nil {
		panic(err)
	}
	fmt.Println(service_resp)

	err = client.DeleteService(service_resp.Id)
	if err != nil {
		panic(err)
	}
}
