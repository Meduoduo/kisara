package client

import (
	"os"

	"github.com/Yeuoly/Takina/src/api"
	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/routine/docker"
	log "github.com/Yeuoly/kisara/src/routine/log"
	"github.com/Yeuoly/kisara/src/types"
)

func launchTakina() {
	takina_token = helper.GetConfigString("takina.token")
	if takina_token == "" {
		log.Panic("[Kisara] Failed to get Takina token from config file")
	}
	var err error
	takina_config, err = os.Open("conf/takina_client.yaml")
	if err != nil {
		log.Panic("[Kisara] Failed to read Takina config file: " + err.Error())
	}

	// stop takina docker daemon if it is running
	api.StopTakinaDockerDaemon()
	// init takina docker daemon
	resp, err := api.InitTakinaDockerDaemon(takina_token, takina_config,
		func(success string) {
		},
		func(err string) {
		},
	)

	takina_container_id = resp.ContainerId

	if err != nil {
		log.Panic("[Kisara] Failed to init Takina Docker Daemon: " + err.Error())
	}
}

func attachTakinaHook() {
	docker.AddOnDockerDaemonStartHook(func(d *docker.Docker, n []types.Network) {
		// wait for docker daemon to be fresh to ensure environment is ready
		<-after_docker_daemon_fresh
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
