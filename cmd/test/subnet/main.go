package main

import (
	"github.com/Yeuoly/kisara/src/helper"
	docker "github.com/Yeuoly/kisara/src/routine/docker"
)

func main() {
	helper.InitServerConfig()
	client := docker.NewDocker()
	_, err := client.CreateNetwork("172.127.0.0/16", "irina-train", true, "bridge")
	if err != nil {
		panic(err)
	}
	//networks, err := client.ListNetwork()
	// if err != nil {
	// 	panic(err)
	// }
	// irina_train_id := ""
	// for _, network := range networks {
	// 	fmt.Println(network.Name)
	// 	if network.Name == "irina-train" {
	// 		irina_train_id = network.ID
	// 	}
	// }
	// err = client.DeleteNetwork(irina_train_id)
	// if err != nil {
	// 	panic(err)
	// }
}
