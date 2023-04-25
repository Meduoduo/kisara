package main

import (
	"fmt"
	"log"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/routine/db"
	"github.com/Yeuoly/kisara/src/types"
)

func main() {
	helper.InitServerConfig()

	db_path := helper.GetConfigString("kisaraClient.db_path")
	if db_path == "" {
		log.Panic("[Kisara] Failed to get Kisara database path from config file")
	}

	// init database, panic if failed
	db.InitKisaraDB(db_path)

	_select()
}

func _select() {
	fmt.Println(db.GetGenericOne[types.DBContainer](
		db.GenericEqual("container_id", "32aa660132df9ba21ca09161eb1d2abaec70b6df545016baf7ba619d0670b0fc"),
	))
}
