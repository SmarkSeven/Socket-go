package route

import (
	"encoding/json"
	"github.com/ylywyn/jpush-api-go-client"
	"log"
	"net"
	"os"
	"push-socket/protocol"
	// "strings"
	"time"
)

const (
	appKey = "0e6202e57d1f1a566dadcb0d"
	secret = "dc0a4eb3deb637e107a8d354"
)

// type Msg struct {
// 	Conditions map[string]string `json:"meta"`
// 	Content    interface{}       `json:"content"`
// }

type Msg struct {
	Conditions map[string]string `json:"meta"`
	Content    PushParam         `json:"content"`
}

// type MsgContent struct {
// 	CoachId     string `json:"coachId"`
// 	StudentId   string `json:"studentId"`
// 	StudentName string `json:"studentName"`
// 	Datetime    string `json:"datetime"`
// }

type MsgContent struct {
	CoachId      string `json:"coachId"`
	StudentId    string `json:"studentId"`
	StudentName  string `json:"studentName"`
	StudentPhone string `json:"studentPhone"`
	Datetime     string `json:"datetime"`
}

type PushResult struct {
	Sendno string `json:"sendno"`
	MsgID  string `json:"msg_id"`
}

type Response struct {
	StatusCode string      `json:"statusCode"`
	Result     interface{} `result`
}

func (this Response) ToBytes() ([]byte, error) {
	content, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func reader(conn net.Conn, readerChannel chan []byte, timeout int) {
	for {
		select {
		case data := <-readerChannel:
			conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
			Business(conn, data)
			break
		case <-time.After(time.Second * time.Duration(timeout)):
			Log("It's really weird to get Nothing!!!")
			conn.Close()
			return
		}
	}
}
func Business(conn net.Conn, data []byte) {
	flag := false
	for _, v := range Routers {
		pred := v[0]
		act := v[1]
		var message Msg
		err := json.Unmarshal(data, &message)
		if err != nil {
			Log(err)
		}
		if pred.(func(entry Msg) bool)(message) {
			result := act.(Controller).Excute(message)
			_, err := WriteResult(conn, result)
			if err != nil {
				Log("conn.WriteResult()", err)
			}
			return
		}
	}
	if !flag {
		_, err := WriteError(conn, "1111", "不能处理此类型的业务")
		if err != nil {
			Log("conn.WriteError()", err)
		}
	}
}

func CheckError(err error) {
	if err != nil {
		log.Printf("Fatal error: %s", err.Error())
		os.Exit(1)
		// return true
	}
}

//长连接
func HandleConnection(conn net.Conn, timeout int) {
	//声明一个临时缓冲区，用来存储被截断的数据
	tmpBuffer := make([]byte, 0)

	//声明一个管道用于接收解包的数据
	readerChannel := make(chan []byte, 12)
	go reader(conn, readerChannel, timeout)

	buffer := make([]byte, 2048)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			Log(conn.RemoteAddr().String(), " connection error: ", err)
			return
		}
		log.Println(n)
		tmpBuffer = protocol.Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
	}

}

func WriteResult(conn net.Conn, result interface{}) (n int, err error) {
	// data, err := json.Marshal(Response{StatusCode: "0000", Result: result})
	// if err != nil {
	// 	return 0, err
	// }
	return conn.Write(result.([]byte))
}

func WriteError(conn net.Conn, statusCode string, result interface{}) (n int, err error) {
	data, err := json.Marshal(Response{StatusCode: statusCode, Result: result})
	if err != nil {
		return 0, err
	}
	return conn.Write(data)
}

func Log(v ...interface{}) {
	log.Println(v...)
}

type Controller interface {
	Excute(message Msg) interface{}
}

var Routers [][2]interface{}

func Route(judge interface{}, controller Controller) {
	switch judge.(type) {
	case func(entry Msg) bool:
		{
			var arr [2]interface{}
			arr[0] = judge
			arr[1] = controller
			Routers = append(Routers, arr)
		}
	case map[string]string:
		{
			defaultJudge := func(entry Msg) bool {
				for keyjudge, valjudge := range judge.(map[string]string) {
					val, ok := entry.Conditions[keyjudge]
					if !ok {
						return false
					}
					if val != valjudge {
						return false
					}
				}
				return true
			}
			var arr [2]interface{}
			arr[0] = defaultJudge
			arr[1] = controller
			Routers = append(Routers, arr)
		}
	default:
		Log("Something is wrong in Router")
	}
}

// 绑定教练
type BindCoach struct {
}

func (this *BindCoach) Excute(message Msg) interface{} {
	// _, err := json.Marshal(message)
	// CheckError(err)
	return PushMsgAndNotice(message)
}

// 预约课程
type BookCourse struct {
}

func (this *BookCourse) Excute(message Msg) interface{} {
	// _, err := json.Marshal(message)
	// CheckError(err)
	return PushMsgAndNotice(message)
}

type PushParam struct {
	CoachId     string `json:"coachId"`
	StudentName string `json:"studentName"`
	Phone       string `json:"phone"`
	// MsgType     string      `json:"msgType"`
	Datetime string                 `json:"datetime"`
	Extra    map[string]interface{} `json:"extras"`
}

func PushMsgAndNotice(message Msg) interface{} {

	log.Printf("MSG_ %#v", message)
	pushParam := message.Content
	var pf jpushclient.Platform
	pf.All()

	var ad jpushclient.Audience
	s := []string{pushParam.CoachId}
	ad.SetAlias(s)

	var notice jpushclient.Notice
	var msg jpushclient.Message

	switch message.Conditions["msgtype"] {
	case "BDJL":
		notice.SetAndroidNotice(&jpushclient.AndroidNotice{
			Alert:  "学员绑定请求",
			Title:  "助驾帮",
			Extras: message.Content.Extra})

		msg.Title = "绑定请求"
		msg.Content = "来自 " + pushParam.StudentName + "[" + pushParam.Phone + "]" + "的绑定申请"
	case "YYKC":
		notice.SetAndroidNotice(&jpushclient.AndroidNotice{
			Alert:  "学员预约了课程",
			Title:  "助驾帮",
			Extras: message.Content.Extra})

		msg.Title = "学员预约了课程"
		msg.Content = "学员:" + pushParam.StudentName + "[" + pushParam.Phone + "]" + "预约了" + pushParam.Datetime + "的课程"
	}
	msg.Extras = message.Content.Extra
	payload := jpushclient.NewPushPayLoad()
	payload.SetPlatform(&pf)
	payload.SetAudience(&ad)
	// payload.SetMessage(&msg)
	payload.SetNotice(&notice)
	bytes, _ := payload.ToBytes()
	//push
	c := jpushclient.NewPushClient(secret, appKey)
	str, err := c.Send(bytes)
	if err != nil {
		log.Println(err)
		data, _ := Response{StatusCode: "1111", Result: err}.ToBytes()
		return data
	} else {
		data, _ := Response{StatusCode: "0000", Result: str}.ToBytes()
		return data
	}
}

// func mirrorHandle(entry Msg) bool {
// 	if entry.Conditions["msgtype"] == "binding" {
// 		return true
// 	}
// 	return false
// }

func init() {
	var bind BindCoach
	var bookCourse BookCourse
	Routers = make([][2]interface{}, 0, 10)
	kvs := make(map[string]string)
	kvs["msgtype"] = "BDJL"
	kvs2 := make(map[string]string)
	kvs2["msgtype"] = "YYKC"
	Route(kvs, &bind)
	Route(kvs2, &bookCourse)
}
