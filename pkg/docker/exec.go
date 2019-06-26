package docker

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/podInteractive/pkg/constant"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

type terminal struct {
	conn        *websocket.Conn
	Address     string
	ContainerId string
	ShellId     string
}

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
type shellResult struct {
	Id string `json:"Id"`
}
type execStartParam struct {
	Detach bool `json:"Detach"`
	Tty    bool `json:"Tty"`
}

func execShellStream(t *terminal) {

	tcpConn, err := net.Dial("tcp", t.Address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tcpConn.Close()
	data := "{\"Tty\":true}"
	dataLength := len([]byte(data))
	body := fmt.Sprintf("POST /exec/%s/start HTTP/1.1\r\nHost: %s\r\nContent-Type: application/json\r\nContent-Length: %s\r\n\r\n%s", t.ShellId, t.Address, fmt.Sprint(dataLength), data)
	_, err = tcpConn.Write([]byte(body))
	if err != nil {
		log.Println(err)
	}

	go func() {
		for {
			bytes := make([]byte, 128)
			_, err := tcpConn.Read(bytes)
			t.conn.WriteMessage(websocket.TextMessage, bytes)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

	}()

	for {
		_, r, err := t.conn.NextReader()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		tcpConn.Write(bytes)
	}
}

func shellId(t *terminal) (string, error) {
	// url := "http://134.44.36.120:2376/containers/f441cd3e0d5f1/exec"
	url := fmt.Sprintf("http://%s/containers/%s/exec", t.Address, t.ContainerId)
	fmt.Println(url)
	param := &execParam{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          constant.DefaultCommand,
		DetachKeys:   "ctrl-p,ctrl-q",
		Privileged:   true,
		Tty:          true,
	}
	request, err := httplib.Post(url).JSONBody(param)
	if err != nil {
		return "", err
	}
	result := &shellResult{}
	err = request.ToJSON(result)
	if err != nil {
		return "", err
	}
	if result.Id == "" {
		return "", errors.New("shellId为空")
	}
	return result.Id, nil
}
func Exec(req *restful.Request, resp *restful.Response) {

	c, err := upgrader.Upgrade(resp, req.Request, nil)
	if err != nil {
		fmt.Println("upgrade:", err)
		return
	}
	defer c.Close()
	t := &terminal{
		conn:        c,
		Address:     "134.44.36.120:2376",
		ContainerId: "f441cd3e0d5f",
	}
	shellId, err := shellId(t)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.ShellId = shellId
	fmt.Println(shellId)
	fmt.Println(t.ShellId)
	fmt.Println(t)
	execShellStream(t)
}
