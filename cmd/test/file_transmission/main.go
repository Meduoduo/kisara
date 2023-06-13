package main

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/gin-gonic/gin"

	_ "embed"
)

//go:embed test.tar.gz
var tarfile []byte

type FileTransmission struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func main() {
	go func() {
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()

		r.POST("/upload", func(c *gin.Context) {
			var fileTransmission FileTransmission
			c.ShouldBind(&fileTransmission)
			fmt.Println(fileTransmission)
			fmt.Println(fileTransmission.File)

			c.JSON(200, gin.H{
				"message": "ok",
			})
		})

		r.Run(":7777")
	}()

	time.Sleep(time.Second * 1)

	fmt.Println("Sending file...")

	// read tarfile
	reader := bytes.NewBuffer(tarfile)

	resp, err := helper.SendPostAndParse[map[string]any](
		"http://localhost:7777/upload",
		helper.HttpPyloadMultipart(
			map[string]string{},
			helper.HttpPayloadMultipartFile("file", reader),
		),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println(resp)
}
