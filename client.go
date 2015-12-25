package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"socket_c/protocol"
	"time"
)

type Msg struct {
	Conditions map[string]string `json:"meta"`
	Content    PushParam         `json:"content"`
}

type PushParam struct {
	CoachId     string                 `json:"coachId"`
	StudentName string                 `json:"studentName"`
	Phone       string                 `json:"phone"`
	Datetime    string                 `json:"datetime"`
	Extra       map[string]interface{} `json:"extras"`
}

type Info struct {
	Id       int64
	PushData string `json:"pushData"`
}
type Response struct {
	StatusCode string `json:"statusCode"`
	Result     string `result`
}

func senderMsg(conn net.Conn) {
	// for i := 0; i < 0; i++ {

	msg := Msg{}
	kvs := make(map[string]string)
	kvs["msgtype"] = "BDJL"
	kvs2 := make(map[string]interface{})
	kvs2["msgtype"] = "BDJL"

	pushParam := PushParam{}
	pushParam.CoachId = "13"
	pushParam.StudentName = "JB"
	pushParam.Phone = "123"
	pushParam.Datetime = "0000-0000"

	pushParam.Extra = kvs2
	msg.Conditions = kvs
	msg.Content = pushParam

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Marchal err %#v", msg)
	}
	conn.Write(protocol.Packet(data))
	buffer := make([]byte, 2048)
	n, err := conn.Read(buffer)
	var message Response
	err = json.Unmarshal(buffer[:n], &message)
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("接收到的数据 %#v", message)
	fmt.Printf("结果 %#v", string(message.Result))
	time.Sleep(1 * time.Second)

}

func main() {
	server := "localhost:6060"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	fmt.Println("connect success")
	senderMsg(conn)

}
