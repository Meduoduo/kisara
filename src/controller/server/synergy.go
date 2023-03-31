package server

import (
	"time"

	"github.com/Yeuoly/kisara/src/controller"
	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/routine/synergy/server"
	"github.com/Yeuoly/kisara/src/types"
	"github.com/gin-gonic/gin"
)

func HandleConnect(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestConnect) {
		var wait_chan = make(chan bool)
		var resp *types.ResponseConnect
		rc.Callback = func(rc types.ResponseConnect) {
			resp = &rc
			wait_chan <- true
		}
		server.AddConnectRequest(rc)
		if !helper.TimeoutWrapper(5*time.Second, wait_chan) {
			r.JSON(200, types.ErrorResponse(-500, "timeout"))
		} else {
			r.JSON(200, types.SuccessResponse(resp))
		}
	})
}

func HandleDisconnect(r *gin.Context) {
	controller.BindRequest(r, func(rd types.RequestDisconnect) {
		server.Disconnect(rd.ClientID)
	})
}

func HandleHeartBeat(r *gin.Context) {
	controller.BindRequest(r, func(rhb types.RequestHeartBeat) {
		server.UpdateHeartBeat(rhb.ClientID)
		r.JSON(200, types.SuccessResponse(types.ResponseHeartBeat{
			ClientID:  rhb.ClientID,
			Timestamp: time.Now().Unix(),
		}))
	})
}

func HandleRecvStatus(r *gin.Context) {
	controller.BindRequest(r, func(rss types.RequestStatus) {
	})
}
