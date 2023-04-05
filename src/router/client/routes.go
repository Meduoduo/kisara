package client

import (
	"github.com/Yeuoly/kisara/src/controller/client"
	"github.com/Yeuoly/kisara/src/router"
	"github.com/gin-gonic/gin"
)

func Setup(eng *gin.Engine) {
	eng.POST(router.URI_CLIENT_CREATE_NETWORK, client.HandleCreateSubnet)
	eng.POST(router.URI_CLIENT_REMOVE_NETWORK, client.HandleDeleteSubnet)
	eng.GET(router.URI_CLIENT_LIST_NETWORK, client.HandleListSubnet)
	eng.GET(router.URI_CLIENT_LIST_CONTAINER, client.HandleListContainer)
	eng.POST(router.URI_CLIENT_LAUNCH_CONTAINER, client.HandleLaunchContainer)
	eng.GET(router.URI_CLIENT_LAUNCH_CONTAINER_CHECK, client.HandleCheckLaunchContainerStatus)
	eng.POST(router.URI_CLIENT_STOP_CONTAINER, client.HandleStopContainer)
	eng.POST(router.URI_CLIENT_REMOVE_CONTAINER, client.HandleRemoveContainer)
	eng.POST(router.URI_CLIENT_EXEC_CONTAINER, client.HandleExecContainer)
	eng.GET(router.URI_CLIENT_LIST_IMAGE, client.HandleListImage)
	eng.POST(router.URI_CLIENT_INSPECT_CONTAINER, client.HandleInspectContainers)
}