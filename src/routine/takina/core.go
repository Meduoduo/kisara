package routine

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/Yeuoly/kisara/src/helper"

	"github.com/aceld/zinx/znet"
)

type TakinaRequest struct {
	Token string `json:"token"`
	Type  string `json:"type"`
	Data  string `json:"data"`
}

type TakinaRequestStartProxy struct {
	ProxyType string `json:"proxy_type"`
	Laddr     string `json:"laddr"`
	Lport     int    `json:"lport"`
}

type TakinaRequestStopProxy struct {
	Laddr string `json:"laddr"`
	Lport int    `json:"lport"`
}

type TakinaResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

type TakinaResponseStartProxy struct {
	Raddr string `json:"raddr"`
	Rport int    `json:"rport"`
}

type TakinaResponseStopProxy struct{}

const (
	TAKINA_TYPE_ADD_PROXY = "add_proxy"
	TAKINA_TYPE_DEL_PROXY = "del_proxy"
	TAKINA_TYPE_GET_PROXY = "get_proxy"
)

var takina_address = ""
var takina_port = 0
var takina_token = ""

func init() {
	takina_address = helper.GetConfigString("takina.address")
	if takina_address == "" {
		log.Panic("takina_address is empty")
	}
	takina_port = helper.GetConfigInteger("takina.port")
	if takina_port == 0 {
		log.Panic("takina_port is empty")
	}
	takina_token = helper.GetConfigString("takina.token")
	if takina_token == "" {
		log.Panic("takina_token is empty")
	}
}

func TakinaRequestAddProxy(laddr string, lport int, protocol string) (string, int, error) {
	request_addproxy, _ := json.Marshal(TakinaRequestStartProxy{
		ProxyType: protocol,
		Laddr:     laddr,
		Lport:     lport,
	})
	request := TakinaRequest{
		Token: takina_token,
		Type:  TAKINA_TYPE_ADD_PROXY,
		Data:  string(request_addproxy),
	}

	conn, err := net.Dial("tcp", takina_address+":"+strconv.Itoa(takina_port))
	if err != nil {
		return "", 0, err
	}

	defer conn.Close()

	text, _ := json.Marshal(request)
	dp := znet.NewDataPack()
	msg, _ := dp.Pack(znet.NewMsgPackage(1, text))
	_, err = conn.Write(msg)
	if err != nil {
		return "", 0, err
	}

	conn.SetDeadline(time.Now().Add(time.Second * 5))
	response_data, err := recvTakina(conn)
	if err != nil {
		return "", 0, err
	}

	response := TakinaResponse{}
	err = json.Unmarshal(response_data, &response)
	if err != nil {
		return "", 0, err
	}

	if response.Code != 0 {
		return "", 0, errors.New(response.Msg)
	}

	response_startproxy := TakinaResponseStartProxy{}
	err = json.Unmarshal([]byte(response.Data), &response_startproxy)
	if err != nil {
		return "", 0, err
	}

	return response_startproxy.Raddr, response_startproxy.Rport, nil
}

func TakinaRequestDelProxy(laddr string, lport int) error {
	request_delproxy, _ := json.Marshal(TakinaRequestStopProxy{
		Laddr: laddr,
		Lport: lport,
	})

	request := TakinaRequest{
		Token: takina_token,
		Type:  TAKINA_TYPE_DEL_PROXY,
		Data:  string(request_delproxy),
	}

	conn, err := net.Dial("tcp", takina_address+":"+strconv.Itoa(takina_port))
	if err != nil {
		return err
	}

	defer conn.Close()

	text, _ := json.Marshal(request)
	dp := znet.NewDataPack()
	msg, _ := dp.Pack(znet.NewMsgPackage(1, text))
	_, err = conn.Write(msg)
	if err != nil {
		return err
	}

	conn.SetDeadline(time.Now().Add(time.Second * 5))
	response_data, err := recvTakina(conn)
	if err != nil {
		return err
	}

	response := TakinaResponse{}
	err = json.Unmarshal(response_data, &response)
	if err != nil {
		return err
	}

	if response.Code != 0 {
		return errors.New(response.Msg)
	}

	return nil
}

func recvTakina(conn net.Conn) ([]byte, error) {
	dp := znet.NewDataPack()

	head_data := make([]byte, dp.GetHeadLen())
	_, err := io.ReadFull(conn, head_data)
	if err != nil {
		return nil, err
	}

	msg_head, err := dp.Unpack(head_data)
	if err != nil {
		return nil, errors.New("unpack head error")
	}

	if msg_head.GetDataLen() > 0 {
		msg := msg_head.(*znet.Message)
		msg.Data = make([]byte, msg.GetDataLen())

		_, err := io.ReadFull(conn, msg.Data)
		if err != nil {
			return nil, err
		}

		return msg.Data, nil
	}

	return []byte{}, nil
}
