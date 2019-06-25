package docker

import (
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type execParam struct {
	AttachStdin  bool     `json:"AttachStdin"`
	AttachStdout bool     `json:"AttachStdout"`
	AttachStderr bool     `json:"AttachStderr"`
	Cmd          []string `json:"cmd"`
	DetachKeys   string   `json:"DetachKeys"`
	Privileged   bool     `json:"Privileged"`
	Tty          bool     `json:"Tty"`
}
type execId struct {
	Id       string   `json:"Id"`
	Warnings []string `json:"Warnings"`
}
type execStartParam struct {
	Detach bool `json:"Detach"`
	Tty    bool `json:"Tty"`
}

func netConn(execid string) {
	conn, err := net.Dial("tcp", "134.44.36.120:2376")
	if err != nil {
		fmt.Println(err)
		return
	}
	data := "{\"Tty\":true}"
	dataLength := len([]byte(data))
	body := fmt.Sprintf("POST /exec/%s/start HTTP/1.1\r\nHost: %s\r\nContent-Type: application/json\r\nContent-Length: %s\r\n\r\n%s", execid, "134.44.36.120:2376", fmt.Sprint(dataLength), data)
	_, err = conn.Write([]byte(body))
	if err != nil {
		log.Println(err)
	}
}
func Exec(req *restful.Request, resp *restful.Response) {

	c, err := upgrader.Upgrade(resp, req.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	url := "http://134.44.36.120:2376/containers/f441cd3e0d5f1/exec"
	//{
	//	"AttachStdin": true,
	//	"AttachStdout": true,
	//	"AttachStderr": true,
	//	"Cmd": ["sh"],
	//	"DetachKeys": "ctrl-p,ctrl-q",
	//	"Privileged": true,
	//	"Tty": true
	//}
	param := &execParam{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/bin/sh"},
		DetachKeys:   "ctrl-p,ctrl-q",
		Privileged:   true,
		Tty:          true,
	}
	request, err := httplib.Post(url).JSONBody(param)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	execId := &execId{}
	err = request.ToJSON(execId)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(execId.Id)

	conn, err := net.Dial("tcp", "134.44.36.120:2376")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	data := "{\"Tty\":true}"
	dataLength := len([]byte(data))
	body := fmt.Sprintf("POST /exec/%s/start HTTP/1.1\r\nHost: %s\r\nContent-Type: application/json\r\nContent-Length: %s\r\n\r\n%s", execId.Id, "134.44.36.120:2376", fmt.Sprint(dataLength), data)
	_, err = conn.Write([]byte(body))
	if err != nil {
		log.Println(err)
	}

	go func() {
		for {

			bytes := make([]byte, 128)
			_, err := conn.Read(bytes)
			c.WriteMessage(websocket.TextMessage, bytes)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

	}()

	for {
		_, r, err := c.NextReader()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		conn.Write(bytes)
	}
}
