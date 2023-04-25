package main

import (
	"fmt"

	"github.com/Yeuoly/kisara/src/routine/docker"
)

func main() {
	list, err := docker.ParseCIDRRange("172.[128-255].[0-255].0/24")
	fmt.Println(list, err)
}
