package client

import (
	"os"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/routine/db"
	log "github.com/Yeuoly/kisara/src/routine/log"
)

var (
	takina_token        string
	takina_config       *os.File
	takina_container_id string
)

func setupConfig() {
	helper.InitServerConfig()

	db_path := helper.GetConfigString("kisaraClient.db_path")
	if db_path == "" {
		log.Panic("[Kisara] Failed to get Kisara database path from config file")
	}

	// init database, panic if failed
	db.InitKisaraDB(db_path)
}
