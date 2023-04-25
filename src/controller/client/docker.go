package client

import (
	"encoding/json"

	"github.com/Yeuoly/kisara/src/controller"
	docker "github.com/Yeuoly/kisara/src/routine/docker"
	log "github.com/Yeuoly/kisara/src/routine/log"
	request "github.com/Yeuoly/kisara/src/routine/request"
	synergy_client "github.com/Yeuoly/kisara/src/routine/synergy/client"
	"github.com/Yeuoly/kisara/src/types"
	"github.com/gin-gonic/gin"
)

func checkClientKey(client_id string, success func() types.KisaraResponse) types.KisaraResponse {
	if client_id == synergy_client.GetClientId() {
		return success()
	}
	return types.ErrorResponse(-403, "Access Deind")
}

func jsonHelperEncoder[T any](obj T) string {
	json, _ := json.Marshal(obj)
	return string(json)
}

func jsonHelperDecoder[T any](text string) T {
	var t T
	json.Unmarshal([]byte(text), &t)
	return t
}

type launchContainerResponseFormat struct {
	Container *types.Container `json:"container"`
	Error     string           `json:"error"`
}

func HandleLaunchContainer(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestLaunchContainer) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseLaunchContainer{}
			response_id := request.CreateNewResponse()
			go func() {
				docker := docker.NewDocker()
				container, err := docker.LaunchContainer(rc.Image, rc.UID, rc.PortProtocol, rc.SubnetName, rc.Module, rc.EnvMount...)
				if err != nil {
					request.FinishRequest(response_id, jsonHelperEncoder(launchContainerResponseFormat{
						Container: nil,
						Error:     err.Error(),
					}))
				} else if container == nil {
					request.FinishRequest(response_id, jsonHelperEncoder(launchContainerResponseFormat{
						Container: nil,
						Error:     "An unexpected error occurred, container is nil",
					}))
				} else {
					request.FinishRequest(response_id, jsonHelperEncoder(launchContainerResponseFormat{
						Container: container,
						Error:     "",
					}))
				}
			}()
			resp.ResponseId = response_id
			resp.ClientID = rc.ClientID
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleCheckLaunchContainerStatus(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestCheckLaunchStatus) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseCheckLaunchStatus{}
			resp.ClientID = rc.ClientID
			response_id := rc.ResponseId
			response_text, ok := request.GetResponse(response_id)
			if !ok {
				resp.Finished = false
				return types.SuccessResponse(resp)
			} else {
				resp.Finished = true
				middleware_response := jsonHelperDecoder[launchContainerResponseFormat](response_text)
				if middleware_response.Error != "" {
					return types.ErrorResponse(-500, middleware_response.Error)
				}
				if middleware_response.Container == nil {
					return types.ErrorResponse(-500, "An unexpected error occurred, container is nil")
				}
				resp.Container = *middleware_response.Container
				return types.SuccessResponse(resp)
			}
		}))
	})
}

func HandleStopContainer(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestStopContainer) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseStopContainer{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			err := docker.StopContainer(rc.ContainerID)
			if err != nil {
				return types.ErrorResponse(-500, err.Error())
			}
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleRemoveContainer(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestRemoveContainer) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseRemoveContainer{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			err := docker.RemoveContainer(rc.ContainerID)
			if err != nil {
				return types.ErrorResponse(-500, err.Error())
			}
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleCreateSubnet(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestCreateNetwork) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseCreateNetwork{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			_, err := docker.CreateNetwork(rc.Subnet, rc.Name, rc.Internal, rc.Driver)
			if err != nil {
				return types.ErrorResponse(-500, err.Error())
			}
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleDeleteSubnet(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestRemoveNetwork) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseRemoveNetwork{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			err := docker.DeleteNetwork(rc.NetworkID)
			if err != nil {
				return types.ErrorResponse(-500, err.Error())
			}
			return types.SuccessResponse(resp)
		}))
	})
}

type pullImageResponseFormat struct {
	Error    string `json:"error"`
	Finished bool   `json:"finished"`
}

func HandlePullImage(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestPullImage) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponsePullImage{}
			message_response_id := request.CreateNewResponse()
			finish_response_id := request.CreateNewResponse()
			go func() {
				log.Info("[PullImage] Pulling image %s", rc.ImageName)
				pull_message_callback := func(message string) {
					log.Info("[PullImage] %s", message)
					request.SetRequestStatusText(message_response_id, message)
				}

				docker := docker.NewDocker()
				image, err := docker.PullImage(rc.ImageName, pull_message_callback)
				if err != nil {
					request.FinishRequest(message_response_id, "Finished (Error)")
					request.FinishRequest(finish_response_id, jsonHelperEncoder(pullImageResponseFormat{
						Error:    err.Error(),
						Finished: true,
					}))
				} else if image == nil {
					request.FinishRequest(message_response_id, "Finished (Image is nil)")
					request.FinishRequest(finish_response_id, jsonHelperEncoder(pullImageResponseFormat{
						Error:    "An unexpected error occurred, image is nil",
						Finished: true,
					}))
				} else {
					request.FinishRequest(message_response_id, "Finished")
					request.FinishRequest(finish_response_id, jsonHelperEncoder(pullImageResponseFormat{
						Error:    "",
						Finished: true,
					}))
				}
			}()
			resp.FinishResponseID = finish_response_id
			resp.MessageResponseId = message_response_id

			resp.ClientID = rc.ClientID
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleCheckPullImage(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestCheckPullImage) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseCheckPullImage{}
			resp.ClientID = rc.ClientID
			message, _ := request.GetResponse(rc.MessageResponseId)
			resp.Message = message
			finish_response_text, finsihed := request.GetResponse(rc.FinishResponseID)
			resp.Finished = finsihed
			if finsihed {
				finish_response := jsonHelperDecoder[pullImageResponseFormat](finish_response_text)
				resp.Error = finish_response.Error
				if resp.Error != "" {
					return types.ErrorResponse(-500, resp.Error)
				}
			}
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleDeleteImage(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestDeleteImage) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseDeleteImage{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			err := docker.DeleteImage(rc.ImageID)
			if err != nil {
				return types.ErrorResponse(-500, err.Error())
			}
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleListImage(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestListImage) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseListImage{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			images, err := docker.ListImage()
			if err != nil {
				return types.ErrorResponse(-500, err.Error())
			}
			if images == nil {
				return types.ErrorResponse(-500, "An unexpected error occurred, images is nil")
			}
			for _, image := range *images {
				if image != nil {
					resp.Images = append(resp.Images, *image)
				}
			}
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleListContainer(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestListContainer) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseListContainer{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			containers, err := docker.ListContainer()
			if err != nil {
				return types.ErrorResponse(-500, err.Error())
			}
			if containers == nil {
				return types.ErrorResponse(-500, "An unexpected error occurred, containers is nil")
			}
			for _, container := range *containers {
				if container != nil {
					resp.Containers = append(resp.Containers, *container)
				}
			}
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleListSubnet(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestListNetwork) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseListNetwork{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			networks, err := docker.ListNetwork()
			if err != nil {
				return types.ErrorResponse(-500, err.Error())
			}
			if networks == nil {
				return types.ErrorResponse(-500, "An unexpected error occurred, networks is nil")
			}
			resp.Networks = networks
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleInspectContainers(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestInspectContainer) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseInspectContainer{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			containers := []types.Container{}
			for _, containerID := range rc.ContainerIDs {
				container, err := docker.InspectContainer(containerID)
				if err != nil {
					return types.ErrorResponse(-500, err.Error())
				}
				if container == nil {
					return types.ErrorResponse(-500, "An unexpected error occurred, container is nil")
				}
				containers = append(containers, *container)
			}
			resp.Containers = containers
			return types.SuccessResponse(resp)
		}))
	})
}

func HandleExecContainer(r *gin.Context) {
	controller.BindRequest(r, func(rc types.RequestExecContainer) {
		r.JSON(200, checkClientKey(rc.ClientID, func() types.KisaraResponse {
			resp := &types.ResponseExecContainer{}
			resp.ClientID = rc.ClientID
			docker := docker.NewDocker()
			err := docker.Exec(rc.ContainerID, rc.Cmd)
			if err != nil {
				return types.ErrorResponse(-500, err.Error())
			}
			return types.SuccessResponse(resp)
		}))
	})
}
