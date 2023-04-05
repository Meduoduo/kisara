package router

const (
	URI_SERVER_CONNECT    = "/connect"    // connect to server
	URI_SERVER_DISCONNECT = "/disconnect" // disconnect from server
	URI_SERVER_HEARTBEAT  = "/heartbeat"  // heartbeat to server
	URI_SERVER_STATUS     = "/status"     // report status to server

	URI_CLIENT_LAUNCH_CONTAINER       = "/container/launch"       // launch container
	URI_CLIENT_LAUNCH_CONTAINER_CHECK = "/container/launch/check" // launch container check
	URI_CLIENT_STOP_CONTAINER         = "/container/stop"         // stop container
	URI_CLIENT_REMOVE_CONTAINER       = "/container/remove"       // remove container
	URI_CLIENT_LIST_CONTAINER         = "/container/list"         // list container
	URI_CLIENT_EXEC_CONTAINER         = "/container/exec"         // exec container
	URI_CLIENT_INSPECT_CONTAINER      = "/container/inspect"      // inspect container
	URI_CLIENT_CREATE_NETWORK         = "/network/create"         // create network
	URI_CLIENT_LIST_NETWORK           = "/network/list"           // list network
	URI_CLIENT_REMOVE_NETWORK         = "/network/remove"         // remove network
	URI_CLIENT_LIST_IMAGE             = "/image/list"             // list image
	URI_CLIENT_PULL_IMAGE             = "/image/pull"             // pull image
	URI_CLIENT_PULL_IMAGE_CHECK       = "/image/pull/check"       // pull image check
	URI_CLIENT_DELETE_IMAGE           = "/image/delete"           // delete image
)
