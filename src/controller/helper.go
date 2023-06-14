package controller

import (
	"fmt"

	"github.com/Yeuoly/kisara/src/types"
	"github.com/gin-gonic/gin"
)

func BindRequest[T any](r *gin.Context, success func(T)) {
	var request T
	var err error
	// check if application/json is set

	context_type := r.GetHeader("Content-Type")
	if context_type == "application/json" {
		err = r.BindJSON(&request)
	} else {
		err = r.Bind(&request)
	}

	if err != nil {
		fmt.Println(request)
		resp := types.ErrorResponse(-400, err.Error())
		r.JSON(200, resp)
		return
	}
	success(request)
}
