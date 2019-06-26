package pod

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/podInteractive/pkg/constant"
	"github.com/weibaohui/podInteractive/pkg/utils"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"net/http"
	"sync"
)

var terminalMaps sync.Map

func saveTerminal(t *terminal) {
	terminalMaps.Store(key(t), t)
}
func removeTerminal(t *terminal) {
	terminalMaps.Delete(key(t))
}
func getTerminal(t *terminal) (*terminal, bool) {
	if value, ok := terminalMaps.Load(key(t)); ok {
		return value.(*terminal), true
	}
	return nil, false
}
func key(t *terminal) string {
	return fmt.Sprintf("%s/%s/%s", t.ns, t.podName, t.containerName)
}

type terminal struct {
	conn          *websocket.Conn
	size          chan *remotecommand.TerminalSize
	ns            string
	podName       string
	containerName string
}

func (t *terminal) Read(p []byte) (n int, err error) {
	_, ps, err := t.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	return copy(p, ps), nil
}
func (t *terminal) Write(p []byte) (n int, err error) {
	writer, err := t.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, nil
	}
	defer writer.Close()
	return writer.Write(p)
}
func (t *terminal) Next() *remotecommand.TerminalSize {
	sizes := <-t.size
	fmt.Println("读取resize", sizes)
	return sizes
}

func Resize(req *restful.Request, resp *restful.Response) {
	t1 := &terminal{
		ns:            "default",
		podName:       "busybox-5b9f476c84-hfz8t",
		containerName: "busybox",
	}
	t, ok := getTerminal(t1)
	if !ok {
		resp.WriteErrorString(500, key(t1)+"没有 Exec 实例")
		return
	}
	size := &struct {
		Width  uint16
		Height uint16
	}{}
	err := req.ReadEntity(size)
	if err != nil {
		resp.WriteErrorString(500, err.Error())
		return
	}
	fmt.Println(size)
	t.size <- &remotecommand.TerminalSize{
		Width:  size.Width,
		Height: size.Height,
	}

}

// 	ws.Route(ws.GET("/ns/{ns}/podName/{podName}/exec")
func PodExec(request *restful.Request, response *restful.Response) {
	params := request.PathParameters()
	ns := params["ns"]
	podName := params["podName"]
	containerName := request.QueryParameter("containerName")
	if containerName == "" {
		// 没有指定，获取第一个
		containerName, _ = GetFirstContainerName(ns, podName)
	}

	c, err := upgrader.Upgrade(response, request.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	t := &terminal{
		conn:          c,
		ns:            ns,
		podName:       podName,
		containerName: containerName,
		size:          make(chan *remotecommand.TerminalSize, 1),
	}
	saveTerminal(t)
	defer removeTerminal(t)

	err = executor(t)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func executor(t *terminal) error {

	req := utils.Cli().CoreV1().RESTClient().Post().
		Resource("pods").
		Name(t.podName).
		Namespace(t.ns).
		SubResource("exec").
		VersionedParams(
			&v1.PodExecOptions{
				TypeMeta:  v12.TypeMeta{},
				Stdin:     true,
				Stdout:    true,
				TTY:       true,
				Container: t.containerName,
				Command:   constant.DefaultCommand,
			},
			scheme.ParameterCodec,
		)
	restConfig, err := clientcmd.BuildConfigFromFlags("", utils.KubeConfigPath())
	if err != nil {
		fmt.Errorf("clientcmd.BuildConfigFromFlags =%s ", err.Error())
		return err
	}
	exec, err := remotecommand.NewSPDYExecutor(restConfig, http.MethodPost, req.URL())
	if err != nil {
		fmt.Errorf("remotecommand.NewSPDYExecutor =%s ", err.Error())
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             t,
		Stdout:            t,
		Stderr:            t,
		Tty:               true,
		TerminalSizeQueue: t,
	})
	return err

}
