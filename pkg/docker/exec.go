package docker

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/podInteractive/pkg/constant"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
)

var terminalMaps sync.Map

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type terminal struct {
	conn        *websocket.Conn
	Address     string
	ContainerId string
	InstanceId  string
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
type execInstance struct {
	Id string `json:"Id"`
}
type execStartParam struct {
	Detach bool `json:"Detach"`
	Tty    bool `json:"Tty"`
}

func saveTerminal(t *terminal) {
	terminalMaps.Store(t.ContainerId, t)
}
func removeTerminal(t *terminal) {
	terminalMaps.Delete(t.ContainerId)
}
func getTerminal(containerId string) (*terminal, bool) {
	if value, ok := terminalMaps.Load(containerId); ok {
		return value.(*terminal), true
	}
	return nil, false
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
	body := fmt.Sprintf("POST /exec/%s/start HTTP/1.1\r\nHost: %s\r\nContent-Type: application/json\r\nContent-Length: %s\r\n\r\n%s", t.InstanceId, t.Address, fmt.Sprint(dataLength), data)
	_, err = tcpConn.Write([]byte(body))
	if err != nil {
		fmt.Println(err)
		return
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

func instanceId(t *terminal) (string, error) {
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
	result := &execInstance{}
	err = request.ToJSON(result)
	if err != nil {
		return "", err
	}
	if result.Id == "" {
		return "", errors.New("exec Instance Id为空")
	}
	return result.Id, nil
}

func Resize(req *restful.Request, resp *restful.Response) {
	///exec/{id}/resize?h=&w=
	containerId := "f441cd3e0d5f"
	t, ok := getTerminal(containerId)
	if !ok {
		resp.WriteErrorString(500, containerId+"没有Exec Instance")
		return
	}
	size := &struct {
		Width  int
		Height int
	}{}
	err := req.ReadEntity(size)
	if err != nil {
		resp.WriteErrorString(500, err.Error())
		return
	}
	url := fmt.Sprintf("http://%s/exec/%s/resize?h=%d&w=%d", t.Address, t.InstanceId, size.Height, size.Width)
	fmt.Println("resize", url)
	_, err = httplib.Post(url).String()
	if err != nil {
		resp.WriteErrorString(500, err.Error())
		return
	}
	resp.WriteAsJson(size)
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
	id, err := instanceId(t)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.InstanceId = id

	saveTerminal(t)
	defer removeTerminal(t)

	//获取exec into container
	execShellStream(t)
}
