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
			resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseCheckLaunchStatus]](
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
	if req.ClientID == "" {
		return types.ResponseListContainer{}, errors.New("client id is empty")
	}

	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseListContainer{}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseListContainer]](
		client.GenerateClientURI(router.URI_CLIENT_LIST_CONTAINER),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseListContainer{}, err
	}

	if resp.Code != 0 {
		return types.ResponseListContainer{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseListContainer{}, errors.New(resp.Data.Error)
	}

	return types.ResponseListContainer{
		ClientID:   client.ClientID,
		Containers: resp.Data.Containers,
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
	if req.ClientID == "" {
		return types.ResponseListNetwork{}, errors.New("client id is empty")
	}

	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseListNetwork{}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.KisaraResponseWrap[types.ResponseListNetwork]](
		client.GenerateClientURI(router.URI_CLIENT_LIST_NETWORK),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)
	if err != nil {
		return types.ResponseListNetwork{}, err
	}

	if resp.Code != 0 {
		return types.ResponseListNetwork{}, errors.New(resp.Message)
	}

	if resp.Data.Error != "" {
		return types.ResponseListNetwork{}, errors.New(resp.Data.Error)
	}

	return types.ResponseListNetwork{
		ClientID: client.ClientID,
		Networks: resp.Data.Networks,
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

func GetNodes() ([]server.ClientItem, error) {
	clients := server.GetNodes()
	return clients, nil
}
