package main

import (
	"fmt"

	"github.com/Yeuoly/kisara/src/api"
)

const test_compose_success = `version: "3.9"
services:
	web:
		image: wulalala:latest
		networks: 
			overlay-default:
			overlay-inner:
		ports:
			- "8080:81"
			- "82"
	controller:
		image: controller:latest
		networks:
			overlay-inner:
	db:
		image: mysql:latest
		networks: 
			overlay-inner:
		ports:
			- "3306"
networks:
	overlay-default:
		ipam:
			driver: overlay
			internal: true
			attachable: true
	overlay-inner:
		ipam:
			driver: overlay
			internal: true
			attachable: true
`

func main() {
	config, err := api.ConvertFromComposeText((test_compose_success), true)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(config.ToYaml())

	// convert back
	compose_config, err := api.ConvertToComposeText(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(compose_config)
}
