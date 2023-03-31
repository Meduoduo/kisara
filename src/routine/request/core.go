package routine

import (
	"fmt"
	"sync"

	"github.com/Yeuoly/kisara/src/helper"
	routine_minitor "github.com/Yeuoly/kisara/src/routine/monitor"
)

type Response struct {
	Text           string
	RemainderTimes int
	Finished       bool
	StatusText     string
}

type RequestField struct {
	Map sync.Map
}

var request_field *RequestField

func (c *RequestField) Init() {
	fmt.Println("[request] request init finsihed")
}

func (c *RequestField) Schedule() {
	//遍历一遍回文列表，对于Times为0且已经完成的请求的，直接进行一个除的删ovo
	request_field.Map.Range(func(key, value interface{}) bool {
		i := value.(*Response)
		if i.RemainderTimes == 0 {
			request_field.Map.Delete(key)
		} else {
			if i.Finished {
				i.RemainderTimes--
			}
		}
		return true
	})
}

func CreateNewResponse() string {
	request_id := helper.Md5(helper.RandomStr(16))
	response := &Response{
		RemainderTimes: 2,
		Finished:       false,
	}
	request_field.Map.Store(request_id, response)
	return request_id
}

//第二个返回参数为当前请求是否已经完成
func GetResponse(request_id string) (string, bool) {
	value, _ := request_field.Map.Load(request_id)
	if value != nil {
		response := value.(*Response)
		if response.Finished {
			return response.Text, true
		} else {
			status_text := response.StatusText
			response.StatusText = ""
			return status_text, false
		}
	}
	return "", false
}

func SetRequestStatusText(request_id string, text string) bool {
	value, _ := request_field.Map.Load(request_id)
	if value != nil {
		response := value.(*Response)
		response.StatusText += text
		return true
	}
	return false
}

func FinishRequest(request_id string, response_text string) {
	value, _ := request_field.Map.Load(request_id)
	if value != nil {
		response := value.(*Response)
		response.Text = response_text
		response.Finished = true
	}
}

func init() {
	request_field = &RequestField{}
	request_field_monitor := routine_minitor.Monitor{
		Type:            routine_minitor.MONITOR_REQUEST,
		RefreshInterval: 30,
		Handler:         request_field,
		Name:            "request-reponse monitor",
		RefreshHandler: func(i interface{}) {
			field := i.(*RequestField)
			field.Schedule()
		},
		InitHandler: func(i interface{}) {
			field := i.(*RequestField)
			field.Init()
		},
	}
	routine_minitor.AppendMonitor(&request_field_monitor)
}
