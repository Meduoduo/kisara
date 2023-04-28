package api

/*
	This package is provide services for server
	It bases on HTTP request to connect to kisara clients
*/

import (
	"errors"
	"time"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/router"
	log "github.com/Yeuoly/kisara/src/routine/log"
	server "github.com/Yeuoly/kisara/src/routine/synergy/server"
	"github.com/Yeuoly/kisara/src/types"
)

func LaunchContainer(req types.RequestLaunchContainer, timeout time.Duration) (types.ResponseFinalLaunchStatus, error) {
	start := time.Now()
	var client types.Client
	var err error
	// if client id is not set, then fetch the lowest demand client
	if req.ClientID == "" {
		client, err = server.FetchLowestDemandClient()
		if err != nil {
			return types.ResponseFinalLaunchStatus{}, err
		}
		req.ClientID = client.ClientID
	} else {
		tmp := server.GetClient(req.ClientID)
		if tmp == nil {
			return types.ResponseFinalLaunchStatus{}, errors.New("client not found")
		}
		client = *tmp
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseLaunchContainer]](
		client.GenerateClientURI(router.URI_CLIENT_LAUNCH_CONTAINER),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseFinalLaunchStatus{}, err
	}

	if resp.Code != 0 {
		return types.ResponseFinalLaunchStatus{}, errors.New(resp.Message)
	}

	response_id := resp.Data.ResponseId
	if response_id == "" {
		return types.ResponseFinalLaunchStatus{}, errors.New("response id is empty, failed to launch container")
	}

	// recycler to check the status of container
	timer := time.NewTimer(timeout - time.Since(start))
	defer timer.Stop()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timer.C:
			return types.ResponseFinalLaunchStatus{}, errors.New("timeout")
		case <-ticker.C:
			resp, err := helper.SendGetAndParse[types.KisaraResponseWrap[types.ResponseCheckLaunchStatus]](
				client.GenerateClientURI(router.URI_CLIENT_LAUNCH_CONTAINER_CHECK),
				helper.HttpTimeout(2000),
				helper.HttpPayloadJson(types.RequestCheckLaunchStatus{
					ClientID:   client.ClientID,
					ResponseId: response_id,
				}),
			)
			if err != nil {
				return types.ResponseFinalLaunchStatus{}, err
			}
			if resp.Code != 0 {
				return types.ResponseFinalLaunchStatus{}, errors.New(resp.Message)
			}
			if resp.Data.ClientID != client.ClientID {
				return types.ResponseFinalLaunchStatus{}, errors.New("mismatched client")
			}
			if resp.Data.Error != "" {
				return types.ResponseFinalLaunchStatus{}, errors.New(resp.Data.Error)
			}
			if !resp.Data.Finished {
				continue
			}
			container := resp.Data.Container
			server.AddContainer(resp.Data.Container.Id, client.ClientID, &container)
			return types.ResponseFinalLaunchStatus{
				ClientID:  client.ClientID,
				Container: resp.Data.Container,
			}, nil
		}
	}
}

func StopContainer(req types.RequestStopContainer, timeout time.Duration) (types.ResponseStopContainer, error) {
	if req.ClientID == "" {
		// try to find the client
		_, client_id, err := server.GetContainer(req.ContainerID)
		if err != nil {
			return types.ResponseStopContainer{}, err
		}
		req.ClientID = client_id
	}

	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseStopContainer{}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseStopContainer]](
		client.GenerateClientURI(router.URI_CLIENT_STOP_CONTAINER),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseStopContainer{}, err
	}

	if resp.Code != 0 {
		return types.ResponseStopContainer{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseStopContainer{}, errors.New(resp.Data.Error)
	}

	return types.ResponseStopContainer{
		ClientID: client.ClientID,
	}, nil
}

func RemoveContainer(req types.RequestRemoveContainer, timeout time.Duration) (types.ResponseRemoveContainer, error) {
	if req.ClientID == "" {
		// try to find the client
		_, client_id, err := server.GetContainer(req.ContainerID)
		if err != nil {
			return types.ResponseRemoveContainer{}, err
		}
		req.ClientID = client_id
	}

	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseRemoveContainer{}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseRemoveContainer]](
		client.GenerateClientURI(router.URI_CLIENT_REMOVE_CONTAINER),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseRemoveContainer{}, err
	}

	if resp.Code != 0 {
		return types.ResponseRemoveContainer{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseRemoveContainer{}, errors.New(resp.Data.Error)
	}

	return types.ResponseRemoveContainer{
		ClientID: client.ClientID,
	}, nil
}

func ListContainer(req types.RequestListContainer, timeout time.Duration) (types.ResponseListContainer, error) {
	clients := []string{}

	if req.ClientID == "" {
		nodes := server.GetNodes()
		for _, node := range nodes {
			clients = append(clients, node.ClientID)
		}
	}

	containers := []types.Container{}

	for _, client_id := range clients {
		client := server.GetClient(client_id)
		if client == nil {
			log.Warn("[Kisara-API] client %s not found", client_id)
			continue
		}

		resp, err := helper.SendGetAndParse[types.KisaraResponseWrap[types.ResponseListContainer]](
			client.GenerateClientURI(router.URI_CLIENT_LIST_CONTAINER),
			helper.HttpTimeout(timeout.Milliseconds()),
			helper.HttpPayloadJson(types.RequestListContainer{
				ClientID: client.ClientID,
			}),
		)
		if err != nil {
			log.Warn("[Kisara-API] client %s list container error: %s", client_id, err.Error())
			continue
		}

		if resp.Code != 0 {
			log.Warn("[Kisara-API] client %s list container error: %s", client_id, resp.Message)
			continue
		}

		if resp.Data.Error != "" {
			log.Warn("[Kisara-API] client %s list container error: %s", client_id, resp.Data.Error)
			continue
		}

		server.FlushContainer(client_id)
		for _, container := range resp.Data.Containers {
			server.AddContainer(container.Id, client_id, &container)
		}

		containers = append(containers, resp.Data.Containers...)
	}

	return types.ResponseListContainer{
		ClientID:   req.ClientID,
		Containers: containers,
	}, nil
}

func ExecContainer(req types.RequestExecContainer, timeout time.Duration) (types.ResponseExecContainer, error) {
	if req.ClientID == "" {
		// try to find the client
		_, client_id, err := server.GetContainer(req.ContainerID)
		if err != nil {
			return types.ResponseExecContainer{}, err
		}
		req.ClientID = client_id
	}

	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseExecContainer{}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseExecContainer]](
		client.GenerateClientURI(router.URI_CLIENT_EXEC_CONTAINER),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseExecContainer{}, err
	}

	if resp.Code != 0 {
		return types.ResponseExecContainer{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseExecContainer{}, errors.New(resp.Data.Error)
	}

	return types.ResponseExecContainer{
		ClientID: client.ClientID,
	}, nil
}

func InspectContainer(req types.RequestInspectContainer, timeout time.Duration) (types.ResponseInspectContainer, error) {
	var node []struct {
		ClientId   string
		Containers []string
	}

	addContainer := func(client_id string, container_id string) {
		for i, n := range node {
			if n.ClientId == client_id {
				node[i].Containers = append(node[i].Containers, container_id)
				return
			}
		}
		node = append(node, struct {
			ClientId   string
			Containers []string
		}{
			ClientId:   client_id,
			Containers: []string{container_id},
		})
	}

	if req.ClientID == "" {
		// try to find the client of each container
		for _, container_id := range req.ContainerIDs {
			_, client_id, err := server.GetContainer(container_id)
			if err != nil {
				continue
			}
			addContainer(client_id, container_id)
		}
	} else {
		for _, container_id := range req.ContainerIDs {
			addContainer(req.ClientID, container_id)
		}
	}

	var containers []types.Container

	for _, n := range node {
		client := server.GetClient(n.ClientId)
		if client == nil {
			log.Warn("[Kisara-API] client %s not found", n.ClientId)
		}

		resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseInspectContainer]](
			client.GenerateClientURI(router.URI_CLIENT_INSPECT_CONTAINER),
			helper.HttpTimeout(timeout.Milliseconds()),
			helper.HttpPayloadJson(types.RequestInspectContainer{
				ClientID:     n.ClientId,
				ContainerIDs: n.Containers,
				HasState:     true,
			}),
		)
		if err != nil {
			log.Warn("[Kisara-API] client %s inspect container error: %s", n.ClientId, err.Error())
			continue
		}

		if resp.Code != 0 {
			log.Warn("[Kisara-API] client %s inspect container error: %s", n.ClientId, resp.Message)
			continue
		}

		if resp.Data.Error != "" {
			log.Warn("[Kisara-API] client %s inspect container error: %s", n.ClientId, resp.Data.Error)
			continue
		}

		containers = append(containers, resp.Data.Containers...)
	}

	return types.ResponseInspectContainer{
		ClientID:   req.ClientID,
		Containers: containers,
	}, nil
}

// create a new network on target node
func CreateNetwork(req types.RequestCreateNetwork, timeout time.Duration) (types.ResponseCreateNetwork, error) {
	if req.ClientID == "" {
		return types.ResponseCreateNetwork{}, errors.New("client id is empty")
	}

	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseCreateNetwork{}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseCreateNetwork]](
		client.GenerateClientURI(router.URI_CLIENT_CREATE_NETWORK),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseCreateNetwork{}, err
	}

	if resp.Code != 0 {
		return types.ResponseCreateNetwork{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseCreateNetwork{}, errors.New(resp.Data.Error)
	}

	return types.ResponseCreateNetwork{
		ClientID: client.ClientID,
	}, nil
}

func ListNetwork(req types.RequestListNetwork, timeout time.Duration) (types.ResponseListNetwork, error) {
	clients := []string{}

	if req.ClientID == "" {
		nodes := server.GetNodes()
		for _, node := range nodes {
			clients = append(clients, node.ClientID)
		}
	}

	networks := []types.Network{}

	for _, client_id := range clients {
		client := server.GetClient(client_id)
		if client == nil {
			log.Warn("[Kisara-API] client %s not found", client_id)
			continue
		}

		resp, err := helper.SendGetAndParse[types.KisaraResponseWrap[types.ResponseListNetwork]](
			client.GenerateClientURI(router.URI_CLIENT_LIST_NETWORK),
			helper.HttpTimeout(timeout.Milliseconds()),
			helper.HttpPayloadJson(types.RequestListNetwork{
				ClientID: client.ClientID,
			}),
		)
		if err != nil {
			log.Warn("[Kisara-API] client %s list network error: %s", client_id, err.Error())
			continue
		}

		if resp.Code != 0 {
			log.Warn("[Kisara-API] client %s list network error: %s", client_id, resp.Message)
			continue
		}

		if resp.Data.Error != "" {
			log.Warn("[Kisara-API] client %s list network error: %s", client_id, resp.Data.Error)
			continue
		}

		networks = append(networks, resp.Data.Networks...)
	}

	return types.ResponseListNetwork{
		ClientID: req.ClientID,
		Networks: networks,
	}, nil
}

func RemoveNetwork(req types.RequestRemoveNetwork, timeout time.Duration) (types.ResponseRemoveNetwork, error) {
	if req.ClientID == "" {
		return types.ResponseRemoveNetwork{}, errors.New("client id is empty")
	}

	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseRemoveNetwork{}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseRemoveNetwork]](
		client.GenerateClientURI(router.URI_CLIENT_REMOVE_NETWORK),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseRemoveNetwork{}, err
	}

	if resp.Code != 0 {
		return types.ResponseRemoveNetwork{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseRemoveNetwork{}, errors.New(resp.Data.Error)
	}

	return types.ResponseRemoveNetwork{
		ClientID: client.ClientID,
	}, nil
}

func ListImage(req types.RequestListImage, timeout time.Duration) (types.ResponseListImage, error) {
	clients := []string{}

	if req.ClientID == "" {
		nodes := server.GetNodes()
		for _, node := range nodes {
			clients = append(clients, node.ClientID)
		}
	}

	images := []types.Image{}

	for _, client_id := range clients {
		client := server.GetClient(client_id)
		if client == nil {
			log.Warn("[Kisara-API] client %s not found", client_id)
			continue
		}

		resp, err := helper.SendGetAndParse[types.KisaraResponseWrap[types.ResponseListImage]](
			client.GenerateClientURI(router.URI_CLIENT_LIST_IMAGE),
			helper.HttpTimeout(timeout.Milliseconds()),
			helper.HttpPayloadJson(types.RequestListImage{
				ClientID: client.ClientID,
			}),
		)
		if err != nil {
			log.Warn("[Kisara-API] client %s list image error: %s", client_id, err.Error())
			continue
		}

		if resp.Code != 0 {
			log.Warn("[Kisara-API] client %s list image error: %s", client_id, resp.Message)
			continue
		}

		if resp.Data.Error != "" {
			log.Warn("[Kisara-API] client %s list image error: %s", client_id, resp.Data.Error)
			continue
		}

		images = append(images, resp.Data.Images...)
	}

	return types.ResponseListImage{
		ClientID: req.ClientID,
		Images:   images,
	}, nil
}

func PullImage(req types.RequestPullImage, timeout time.Duration, message_callback func(string)) (types.ResponseFinalPullImageStatus, error) {
	if req.ClientID == "" {
		return types.ResponseFinalPullImageStatus{}, errors.New("client id is empty")
	}

	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseFinalPullImageStatus{
			ClientID: req.ClientID,
			Error:    "client not found",
		}, errors.New("client not found")
	}

	start := time.Now()

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponsePullImage]](
		client.GenerateClientURI(router.URI_CLIENT_PULL_IMAGE),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseFinalPullImageStatus{
			ClientID: req.ClientID,
			Error:    err.Error(),
		}, err
	}

	if resp.Code != 0 {
		return types.ResponseFinalPullImageStatus{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseFinalPullImageStatus{
			ClientID: req.ClientID,
			Error:    resp.Data.Error,
		}, errors.New(resp.Data.Error)
	}

	// recycler to check the status of container
	timer := time.NewTimer(timeout - time.Since(start))
	defer timer.Stop()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timer.C:
			return types.ResponseFinalPullImageStatus{}, errors.New("timeout")
		case <-ticker.C:
			resp, err := helper.SendGetAndParse[types.KisaraResponseWrap[types.ResponseCheckPullImage]](
				client.GenerateClientURI(router.URI_CLIENT_PULL_IMAGE_CHECK),
				helper.HttpTimeout(2000),
				helper.HttpPayloadJson(types.RequestCheckPullImage{
					ClientID:          client.ClientID,
					MessageResponseId: resp.Data.MessageResponseId,
					FinishResponseID:  resp.Data.FinishResponseID,
				}),
			)
			if err != nil {
				return types.ResponseFinalPullImageStatus{}, err
			}
			if resp.Code != 0 {
				return types.ResponseFinalPullImageStatus{}, errors.New(resp.Message)
			}
			if resp.Data.ClientID != client.ClientID {
				return types.ResponseFinalPullImageStatus{}, errors.New("client id not match")
			}
			if resp.Data.Error != "" {
				return types.ResponseFinalPullImageStatus{}, errors.New(resp.Data.Error)
			}
			if resp.Data.Message != "" {
				message_callback(resp.Data.Message)
			}
			if !resp.Data.Finished {
				continue
			}
			return types.ResponseFinalPullImageStatus{
				ClientID: req.ClientID,
				Error:    "",
			}, nil
		}
	}
}

func DeleteImage(req types.RequestDeleteImage, timeout time.Duration) (types.ResponseDeleteImage, error) {
	if req.ClientID == "" {
		return types.ResponseDeleteImage{}, errors.New("client id is empty")
	}

	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseDeleteImage{
			ClientID: req.ClientID,
			Error:    "client not found",
		}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseDeleteImage]](
		client.GenerateClientURI(router.URI_CLIENT_DELETE_IMAGE),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseDeleteImage{
			ClientID: req.ClientID,
			Error:    err.Error(),
		}, err
	}

	if resp.Code != 0 {
		return types.ResponseDeleteImage{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseDeleteImage{
			ClientID: req.ClientID,
			Error:    resp.Data.Error,
		}, errors.New(resp.Data.Error)
	}

	return types.ResponseDeleteImage{
		ClientID: req.ClientID,
	}, nil
}

func GetNodes() ([]server.ClientItem, error) {
	clients := server.GetNodes()
	return clients, nil
}

func LaunchService(req types.RequestLaunchService, message_callback func(string), timeout time.Duration) (types.ResponseFinalLaunchServiceStatus, error) {
	start := time.Now()
	var client types.Client
	var err error
	// if client id is not set, then fetch the lowest demand client
	if req.ClientID == "" {
		client, err = server.FetchLowestDemandClient()
		if err != nil {
			return types.ResponseFinalLaunchServiceStatus{}, err
		}
		req.ClientID = client.ClientID
	} else {
		tmp := server.GetClient(req.ClientID)
		if tmp == nil {
			return types.ResponseFinalLaunchServiceStatus{}, errors.New("client not found")
		}
		client = *tmp
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseLaunchService]](
		client.GenerateClientURI(router.URI_CLIENT_LAUNCH_SERVICE),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseFinalLaunchServiceStatus{}, err
	}

	if resp.Code != 0 {
		return types.ResponseFinalLaunchServiceStatus{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseFinalLaunchServiceStatus{}, errors.New(resp.Data.Error)
	}

	message_response_id := resp.Data.MessageResponseId
	if message_response_id == "" {
		return types.ResponseFinalLaunchServiceStatus{}, errors.New("response id is empty, failed to launch service")
	}
	finish_response_id := resp.Data.FinishResponseID
	if finish_response_id == "" {
		return types.ResponseFinalLaunchServiceStatus{}, errors.New("finish response id is empty, failed to launch service")
	}

	// recycler to check the status of container
	timer := time.NewTimer(timeout - time.Since(start))
	defer timer.Stop()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timer.C:
			return types.ResponseFinalLaunchServiceStatus{}, errors.New("timeout")
		case <-ticker.C:
			resp, err := helper.SendGetAndParse[types.KisaraResponseWrap[types.ResponseCheckLaunchService]](
				client.GenerateClientURI(router.URI_CLIENT_LAUNCH_SERVICE_CHECK),
				helper.HttpTimeout(2000),
				helper.HttpPayloadJson(types.RequestCheckLaunchService{
					ClientID:          client.ClientID,
					MessageResponseId: message_response_id,
					FinishResponseID:  finish_response_id,
				}),
			)
			if err != nil {
				return types.ResponseFinalLaunchServiceStatus{}, err
			}
			if resp.Code != 0 {
				return types.ResponseFinalLaunchServiceStatus{}, errors.New(resp.Message)
			}
			if resp.Data.ClientID != client.ClientID {
				return types.ResponseFinalLaunchServiceStatus{}, errors.New("mismatched client")
			}
			if resp.Data.Error != "" {
				return types.ResponseFinalLaunchServiceStatus{}, errors.New(resp.Data.Error)
			}
			message_callback(resp.Data.Message)
			if !resp.Data.Finished {
				continue
			}
			service := resp.Data.Service
			for _, v := range service.Containers {
				server.AddContainer(v.Id, client.ClientID, &v)
			}
			server.AddService(service.Id, client.ClientID, &service)
			return types.ResponseFinalLaunchServiceStatus{
				ClientID: client.ClientID,
				Service:  resp.Data.Service,
			}, nil
		}
	}
}

func StopService(req types.RequestStopService, timeout time.Duration) (types.ResponseStopContainer, error) {
	start := time.Now()
	var client types.Client
	var err error
	// if client id is not set, then fetch the lowest demand client
	if req.ClientID == "" {
		// try to find the client
		_, client_id, err := server.GetService(req.ServiceID)
		if err != nil {
			return types.ResponseStopContainer{}, err
		}
		req.ClientID = client_id
		client_temp := server.GetClient(client_id)
		if client_temp == nil {
			return types.ResponseStopContainer{}, errors.New("client not found")
		}

		client = *client_temp
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseStopService]](
		client.GenerateClientURI(router.URI_CLIENT_STOP_SERVICE),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseStopContainer{}, err
	}

	if resp.Code != 0 {
		return types.ResponseStopContainer{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseStopContainer{}, errors.New(resp.Data.Error)
	}

	response_id := resp.Data.ResponseID
	if response_id == "" {
		return types.ResponseStopContainer{}, errors.New("response id is empty, failed to launch service")
	}

	// recycler to check the status of container
	timer := time.NewTimer(timeout - time.Since(start))
	defer timer.Stop()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timer.C:
			return types.ResponseStopContainer{}, errors.New("timeout")
		case <-ticker.C:
			resp, err := helper.SendGetAndParse[types.KisaraResponseWrap[types.ResponseCheckStopService]](
				client.GenerateClientURI(router.URI_CLIENT_STOP_SERVICE_CHECK),
				helper.HttpTimeout(2000),
				helper.HttpPayloadJson(types.RequestCheckStopService{
					ClientID:   client.ClientID,
					ResponseID: response_id,
				}),
			)
			if err != nil {
				return types.ResponseStopContainer{}, err
			}
			if resp.Code != 0 {
				return types.ResponseStopContainer{}, errors.New(resp.Message)
			}
			if resp.Data.ClientID != client.ClientID {
				return types.ResponseStopContainer{}, errors.New("mismatched client")
			}
			if resp.Data.Error != "" {
				return types.ResponseStopContainer{}, errors.New(resp.Data.Error)
			}
			if !resp.Data.Finished {
				continue
			}
			service, _, err := server.GetService(req.ServiceID)
			if err != nil {
				return types.ResponseStopContainer{}, err
			}
			for _, v := range service.Containers {
				server.DeleteContainer(v.Id)
			}
			return types.ResponseStopContainer{
				ClientID: client.ClientID,
				Error:    "",
			}, nil
		}
	}
}

func ListServices(req types.RequestListService, timeout time.Duration) (types.ResponseListService, error) {
	clients := []string{}

	if req.ClientID == "" {
		nodes := server.GetNodes()
		for _, node := range nodes {
			clients = append(clients, node.ClientID)
		}
	}

	containers := []types.Service{}
	for _, client_id := range clients {
		client := server.GetClient(client_id)
		if client == nil {
			log.Warn("[Kisara-API] client %s not found", client_id)
			continue
		}

		req.ClientID = client_id
		resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseListService]](
			client.GenerateClientURI(router.URI_CLIENT_LIST_SERVICE),
			helper.HttpTimeout(timeout.Milliseconds()),
			helper.HttpPayloadJson(req),
		)
		if err != nil {
			return types.ResponseListService{}, err
		}
		if resp.Code != 0 {
			return types.ResponseListService{}, errors.New(resp.Message)
		}
		if resp.Data.Error != "" {
			return types.ResponseListService{}, errors.New(resp.Data.Error)
		}
		containers = append(containers, resp.Data.Services...)
	}

	return types.ResponseListService{
		Services: containers,
	}, nil
}
