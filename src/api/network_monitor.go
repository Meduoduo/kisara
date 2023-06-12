package api

import (
	"errors"
	"time"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/router"
	"github.com/Yeuoly/kisara/src/routine/synergy/server"
	"github.com/Yeuoly/kisara/src/types"
)

func RunNetworkMonitor(req types.RequestNetworkMonitorRun, timeout time.Duration, message_callback func(string)) (types.ResponseFinalNetworkMonitorStatus, error) {
	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseFinalNetworkMonitorStatus{}, errors.New("client not found")
	}

	if req.Context == nil {
		return types.ResponseFinalNetworkMonitorStatus{}, errors.New("context is nil")
	}

	start := time.Now()
	resp, err := helper.SendPostAndParse[types.ResponseNetworkMonitorRun](
		client.GenerateClientURI(router.URI_CLIENT_NETWORK_MONITOR_RUN),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPyloadMultipart(map[string]string{
			"client_id":    req.ClientID,
			"network_name": req.NetworkName,
		}, helper.HttpPayloadMultipartFile("context", *req.Context)),
	)

	if err != nil {
		return types.ResponseFinalNetworkMonitorStatus{}, err
	}

	if resp.Error != "" {
		return types.ResponseFinalNetworkMonitorStatus{}, errors.New(resp.Error)
	}

	if time.Since(start) < time.Millisecond*10 {
		return types.ResponseFinalNetworkMonitorStatus{}, errors.New("timeout")
	}

	response_id := resp.ResponseId
	finish_response_id := resp.FinishResponseID

	timeout_timer := time.NewTimer(timeout)
	cycle_tick := time.NewTicker(time.Second * 1)
	defer timeout_timer.Stop()
	defer cycle_tick.Stop()

	for {
		select {
		case <-timeout_timer.C:
			return types.ResponseFinalNetworkMonitorStatus{}, errors.New("timeout")
		case <-cycle_tick.C:
			resp, err := helper.SendGetAndParse[types.ResponseNetworkMonitorCheck](
				client.GenerateClientURI(router.URI_CLIENT_NETWORK_MONITOR_RUN_CHECK),
				helper.HttpTimeout((timeout - time.Since(start)).Milliseconds()),
				helper.HttpPayloadJson(types.RequestNetworkMonitorCheck{
					ClientID:         req.ClientID,
					ResponseId:       response_id,
					FinishResponseID: finish_response_id,
				}),
			)

			if err != nil {
				return types.ResponseFinalNetworkMonitorStatus{}, err
			}

			message_callback(resp.Message)
			if resp.Finished {
				if resp.Error != "" {
					return types.ResponseFinalNetworkMonitorStatus{}, errors.New(resp.Error)
				}

				return types.ResponseFinalNetworkMonitorStatus{
					ClientID:                  req.ClientID,
					Error:                     resp.Error,
					NetworkMonitorContainerId: resp.NetworkMonitorContainerId,
				}, nil
			}
		}
	}
}

func StopNetworkMonitor(req types.RequestNetworkMonitorStop, timeout time.Duration) (types.ResponseNetworkMonitorStop, error) {
	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseNetworkMonitorStop{}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.ResponseNetworkMonitorStop](
		client.GenerateClientURI(router.URI_CLIENT_NETWORK_MONITOR_STOP),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)

	if err != nil {
		return types.ResponseNetworkMonitorStop{}, err
	}

	if resp.Error != "" {
		return types.ResponseNetworkMonitorStop{}, errors.New(resp.Error)
	}

	return resp, nil
}

func RunNetworkMonitorScript(req types.RequestNetworkMonitorRunScript, timeout time.Duration) (types.ResponseNetworkMonitorRunScript, error) {
	client := server.GetClient(req.ClientID)
	if client == nil {
		return types.ResponseNetworkMonitorRunScript{}, errors.New("client not found")
	}

	resp, err := helper.SendPostAndParse[types.ResponseNetworkMonitorRunScript](
		client.GenerateClientURI(router.URI_CLIENT_NETWORK_MONITOR_SCRIPT),
		helper.HttpTimeout(timeout.Milliseconds()),
		helper.HttpPayloadJson(req),
	)

	if err != nil {
		return types.ResponseNetworkMonitorRunScript{}, err
	}

	if resp.Error != "" {
		return types.ResponseNetworkMonitorRunScript{}, errors.New(resp.Error)
	}

	return resp, nil
}
