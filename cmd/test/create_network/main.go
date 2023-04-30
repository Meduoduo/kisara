package main

import (
	"fmt"
	"time"

	"github.com/Yeuoly/kisara/src/client"
	"github.com/Yeuoly/kisara/src/routine/docker"
	"github.com/Yeuoly/kisara/src/types"
	uuid "github.com/satori/go.uuid"
)

/*
	2023/04/30 there has to be a problem with the network creation
	docker network creation seems to be a sync operation, which means
	we have to wait for the network to be created before we can create
	the containers, this package is created to test the network creation
*/

func main() {
	go client.Main()
	time.Sleep(time.Second * 20)
	testService()

	select {}
}

func testService() {
	service_json := `{"total_score":1300,"network_count":3,"container_count":4,"containers":[{"image":"yeuoly/service_dirty_data:v1","ports":[{"port":80,"protocol":"tcp"}],"networks":[{"network":"A","random_cidr":true},{"network":"B","random_cidr":true}],"flags":[{"flag_command":"echo $flag > /tmp/flag","flag_score":100,"flag_uuid":"74d7ffa2-e701-4cab-823e-932124f860c4"}],"env":{}},{"image":"yeuoly/service_dirtydata_auth:v1","ports":[],"networks":[{"network":"A","random_cidr":true},{"network":"C","random_cidr":true}],"flags":[{"flag_command":"sleep 10 && mysql -uuser -pawdawddasdsadsa auth -e \"UPDATE flag SET flag='$flag'\"","flag_score":300,"flag_uuid":"6b8fb8b9-4713-4d23-b45e-c1a2637ab2b5"}],"env":{}},{"image":"yeuoly/service_dirtydata_ssh:v1","ports":[],"networks":[{"network":"B","random_cidr":true}],"flags":[{"flag_command":"echo $flag > /flag","flag_score":300,"flag_uuid":"5e825514-a205-4e82-ba8f-b0a144a2f525"},{"flag_command":"echo $flag > /tmp/flag && chmod 000 /tmp/flag","flag_score":300,"flag_uuid":"c2f313ed-bc79-44dd-91f7-cf317102ed01"}],"env":{}},{"image":"yeuoly/service_dirtydata_db:v1","ports":[],"networks":[{"network":"C","random_cidr":true}],"flags":[{"flag_command":"mysql -uroot -proot leak -e \"UPDATE flag SET flag = '$flag' WHERE id = 5001\";","flag_score":300,"flag_uuid":"b2418968-98f7-434d-a2ae-d9575eb98f90"}],"env":{}}]}`
	service := types.KisaraService{
		Id:          uuid.NewV4().String(),
		Name:        "service-test",
		Description: "service-test",
		Owner:       9,
		Config:      string(service_json),
	}

	client := docker.NewDocker()
	defer client.Stop()

	for i := 0; i < 100; i++ {
		service_resp, err := client.CreateService(service)
		if err != nil {
			panic(err)
		}
		fmt.Println(service_resp)

		err = client.DeleteService(service_resp.Id)
		if err != nil {
			panic(err)
		}
	}
}
