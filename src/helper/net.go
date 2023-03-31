package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
)

func GetAvaliablePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)

	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

//https request post json data
func HttpPostJson(url string, data interface{}) (int, string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return -1, "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return -1, "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, "", err
	}

	return resp.StatusCode, string(body), nil
}

//https request get, data is a struct which will be converted to urlencoded query string
//example:
// type User struct {
// 	Name string `json:"name"`
// }
//
// httpsGet("http://www.example.com?name=hello", &User{})
func HttpGet(url string, data interface{}) (int, string, error) {
	//convert struct { a: b } to ?a=b use reflect
	queryString := ""
	if data != nil {
		queryString = "?"
		//get struct type
		t := reflect.TypeOf(data)
		//get struct value
		v := reflect.ValueOf(data)
		//get struct field count
		for i := 0; i < t.NumField(); i++ {
			//get struct field
			field := t.Field(i)
			//get struct field tag
			tag := field.Tag.Get("json")
			//get struct field value
			value := v.Field(i).Interface()
			//convert value to string
			valueString := fmt.Sprintf("%v", value)
			//append to query string
			queryString += tag + "=" + valueString + "&"
		}
		//remove last &
		queryString = queryString[:len(queryString)-1]
	}

	resp, err := http.Get(url + "?" + queryString)
	if err != nil {
		return -1, "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, "", err
	}

	return resp.StatusCode, string(body), nil
}
