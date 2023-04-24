package client

import (
	"github.com/Yeuoly/kisara/src/routine/docker"
	log "github.com/Yeuoly/kisara/src/routine/log"
	"github.com/Yeuoly/kisara/src/types"
)

func attachTakinaHook() {
	docker.AddOnDockerDaemonStartHook(func(d *docker.Docker, n []types.Network) {
		// connect takina container to all networks
		log.Info("[Kisara] Connecting Takina container to all networks")
		for _, network := range n {
			err := d.ConnectContainerToNetwork(takina_container_id, network.Id)
			if err != nil {
				log.Error("[Kisara] Failed to connect Takina container to network: " + err.Error())
			}
		}
	})

	docker.AddBeforeNetworkRemoveHook(func(c *docker.Docker, network types.Network) {
		// remove takina container from network
		log.Info("[Kisara] Disconnecting Takina container from network %s", network.Name)
		err := c.DisconnectContainerFromNetwork(takina_container_id, network.Id)
		if err != nil {
			log.Error("[Kisara] Failed to disconnect Takina container from network: " + err.Error())
		}
	})

	docker.AddOnNetworkCreateHook(func(d *docker.Docker, n types.Network) {
		// connect takina container to network
		log.Info("[Kisara] Connecting Takina container to network %s", n.Name)
		err := d.ConnectContainerToNetwork(takina_container_id, n.Id)
		if err != nil {
			log.Error("[Kisara] Failed to connect Takina container to network: " + err.Error())
		}
	})
}
