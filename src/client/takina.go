package client

import (
	"errors"
	"os"
	"time"

	"github.com/Yeuoly/Takina/src/api"
	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/routine/docker"
	log "github.com/Yeuoly/kisara/src/routine/log"
	kisara_types "github.com/Yeuoly/kisara/src/types"
	"github.com/docker/docker/api/types"
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
	docker.AddOnDockerDaemonStartHook(func(d *docker.Docker, n []kisara_types.Network) {
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

	docker.AddBeforeNetworkRemoveHook(func(c *docker.Docker, network kisara_types.Network) error {
		// remove takina container from network
		log.Info("[Kisara] Disconnecting Takina container from network %s", network.Name)
		err := c.DisconnectContainerFromNetwork(takina_container_id, network.Id)
		if err != nil {
			log.Error("[Kisara] Failed to disconnect Takina container from network: " + err.Error())
			return err
		}

		// check if takina is removed from network already
		for i := 0; i < 30; i++ {
			// inspect network
			network, err := c.Client.NetworkInspect(*c.Ctx, network.Id, types.NetworkInspectOptions{})
			if err != nil {
				log.Error("[Kisara] Failed to inspect network: " + err.Error())
				return err
			}

			// check if takina is removed from network
			found_takina := false
			for _, container := range network.Containers {
				if container.Name == "takina" {
					found_takina = true
					break
				}
			}

			if !found_takina {
				// takina is removed from network
				log.Info("[Kisara] Disconnect Takina container from network %s successfully", network.Name)
				return nil
			}

			// wait for 1 second
			time.Sleep(time.Second)
		}

		// network is not removed
		return errors.New("network is not removed")
	})

	docker.AddOnNetworkCreateHook(func(d *docker.Docker, n kisara_types.Network) {
		// connect takina container to network
		log.Info("[Kisara] Connecting Takina container to network %s", n.Name)
		err := d.ConnectContainerToNetwork(takina_container_id, n.Id)
		if err != nil {
			log.Error("[Kisara] Failed to connect Takina container to network: " + err.Error())
		}
	})
}
