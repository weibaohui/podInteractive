package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/podInteractive/pkg/constant"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/tools/remotecommand"
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
	size        chan *remotecommand.TerminalSize
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
type instance struct {
	Id string `json:"Id"`
}

func (t *terminal) saveTerminal() {
	terminalMaps.Store(t.ContainerId, t)
}
func (t *terminal) removeTerminal() {
	terminalMaps.Delete(t.ContainerId)
}
func getTerminal(containerId string) (*terminal, bool) {
	if value, ok := terminalMaps.Load(containerId); ok {
		return value.(*terminal), true
	}
	return nil, false
}
func (t *terminal) execShellStream(ctx context.Context, eg *errgroup.Group) {

	eg.Go(func() error {
		tcpConn, err := net.Dial("tcp", t.Address)
		if err != nil {
			return err
		}
		defer tcpConn.Close()
		data := "{\"Tty\":true}"
		dataLength := len([]byte(data))
		body := fmt.Sprintf("POST /exec/%s/start HTTP/1.1\r\nHost: %s\r\nContent-Type: application/json\r\nContent-Length: %s\r\n\r\n%s", t.InstanceId, t.Address, fmt.Sprint(dataLength), data)
		_, err = tcpConn.Write([]byte(body))
		if err != nil {
			return err
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
			_, bytes, err := t.conn.ReadMessage()
			if err != nil {
				return err
			}
			size := remotecommand.TerminalSize{}
			if err = json.Unmarshal(bytes, &size); err == nil {
				t.size <- &size
			} else {
				tcpConn.Write(bytes)
			}
		}
	})

}

func (t *terminal) instanceId() (string, error) {
	// url := "http://134.44.36.120:2376/containers/f441cd3e0d5f1/exec"
	url := fmt.Sprintf("http://%s/containers/%s/exec", t.Address, t.ContainerId)
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
	result := &instance{}
	err = request.ToJSON(result)
	if err != nil {
		return "", err
	}
	if result.Id == "" {
		return "", errors.New("exec Instance Id为空")
	}
	return result.Id, nil
}

func (t *terminal) resizeFromWs(ctx context.Context, eg *errgroup.Group) {
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case size := <-t.size:
				doResize(t.ContainerId, size)
			}
		}
	})
}
func doResize(containerId string, size *remotecommand.TerminalSize) error {
	t, ok := getTerminal(containerId)
	if !ok {
		return errors.New(containerId + "没有Exec Instance")
	}
	url := fmt.Sprintf("http://%s/exec/%s/resize?h=%d&w=%d", t.Address, t.InstanceId, size.Height, size.Width)
	_, err := httplib.Post(url).String()
	if err != nil {
		return err
	}
	return nil
}

func Resize(req *restful.Request, resp *restful.Response) {
	containerId := req.QueryParameter("containerId")
	size := &remotecommand.TerminalSize{}
	err := req.ReadEntity(size)
	if err != nil {
		resp.WriteErrorString(500, err.Error())
		return
	}
	err = doResize(containerId, size)
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
	cancelCtx, cancel := context.WithCancel(req.Request.Context())
	eg, ctx := errgroup.WithContext(cancelCtx)

	t := &terminal{
		conn:        c,
		Address:     "134.44.36.120:2376",
		ContainerId: req.PathParameter("containerId"),
		size:        make(chan *remotecommand.TerminalSize, 1),
	}
	id, err := t.instanceId()
	if err != nil {
		fmt.Println(err)
		return
	}
	t.InstanceId = id

	t.saveTerminal()
	defer t.removeTerminal()

	t.resizeFromWs(ctx, eg)
	//获取exec into container
	t.execShellStream(ctx, eg)

	err = eg.Wait()
	if err != nil {
		cancel()
	}
}
