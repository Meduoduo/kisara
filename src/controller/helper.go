package controller

import (
	"github.com/Yeuoly/kisara/src/types"
	"github.com/gin-gonic/gin"
)

func BindRequest[T any](r *gin.Context, success func(T)) {
	var request T
	err := r.ShouldBind(&request)
	if err != nil {
		resp := types.ErrorResponse(-400, err.Error())
		r.JSON(200, resp)
		return
	}
	success(request)
}
