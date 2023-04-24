package client

import (
	"os"

	"github.com/Yeuoly/Takina/src/api"
	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/routine/db"
	log "github.com/Yeuoly/kisara/src/routine/log"
)

var (
	takina_token        string
	takina_config       *os.File
	takina_container_id string
)

func setup() {
	helper.InitServerConfig()

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

	db_path := helper.GetConfigString("kisaraClient.db_path")
	if db_path == "" {
		log.Panic("[Kisara] Failed to get Kisara database path from config file")
	}

	// init database, panic if failed
	db.InitKisaraDB(db_path)
}
