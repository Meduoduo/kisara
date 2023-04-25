package client

import (
	"github.com/Yeuoly/kisara/src/routine/docker"
	log "github.com/Yeuoly/kisara/src/routine/log"
)

var after_docker_daemon_fresh = make(chan bool, 1)

func initDocker(cidr_expression string) {
	// refresh docker daemon environment
	docker.InitDocker()

	cidrs, err := docker.InitCIDRPool(cidr_expression)
	if err != nil {
		log.Panic("[Kisara] Failed to init CIDR pool: " + err.Error())
	}
	log.Info("[Kisara] CIDR pool initialized: get %d CIDRs", cidrs)

	after_docker_daemon_fresh <- true
}
