package pod

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/podInteractive/pkg"
	"golang.org/x/sync/errgroup"
	"html/template"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Home(request *restful.Request, response *restful.Response) {
	template, err := template.ParseFiles("./view/container_log.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	template.Execute(response, request.Request.Host)
}
func Exec(request *restful.Request, response *restful.Response) {
	template, err := template.ParseFiles("./view/container_exec.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	template.Execute(response, request.Request.Host)
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
        document.getElementById("msg_end").scrollIntoView();

    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
		var ns = document.getElementById("ns").value;
        var podName = document.getElementById("podName").value;
    	var containerName = document.getElementById("containerName").value;
		var path="ws://{{.}}/ns/"+ns+"/podName/"+podName+"/log?containerName="+containerName
        ws = new WebSocket(path);
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print( evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
       output.innerHTML="";
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">

<form>
<button id="open">获取日志</button>
<button id="close">关闭</button>
<p>ns:<input id="ns" type="text" value="default">
<p>podName<input id="podName" type="text" value="podName">
<p>containerName<input id="containerName" type="text" value="">
</form>
</td><td valign="top" width="50%">
</td></tr></table>
<div id="output"></div>
<div id="msg_end" style="height:0px; overflow:hidden"></div>

</body>
</html>
`))

// 	ws.Route(ws.GET("/ns/{ns}/podName/{podName}/log")
func GetContainerLog(request *restful.Request, response *restful.Response) {
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
	cancelCtx, cancel := context.WithCancel(request.Request.Context())
	readerGroup, ctx := errgroup.WithContext(cancelCtx)

	go func() {
		for {
			if _, _, err := c.NextReader(); err != nil {
				cancel()
				c.Close()
				break
			}
		}
	}()
	logEvent := make(chan []byte)

	ReadLog(ctx, readerGroup, logEvent, ns, podName, containerName)

	go func() {
		readerGroup.Wait()
		close(logEvent)
	}()
	done := false
	for !done {
		select {
		case item, ok := <-logEvent:
			if !ok {
				done = true
				break
			}
			if err := writeData(c, item); err != nil {
				cancel()
			}

		}
	}

}

func writeData(c *websocket.Conn, buf []byte) error {
	messageWriter, err := c.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	if _, err := messageWriter.Write(buf); err != nil {
		return err
	}
	return messageWriter.Close()
}
func GetFirstContainerName(ns string, podName string) (string, error) {
	pod, err := pkg.Cli().CoreV1().Pods(ns).Get(podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if len(pod.Spec.Containers) == 0 {
		return "", errors.New("没有容器")
	}
	return pod.Spec.Containers[0].Name, nil
}

func ReadLog(ctx context.Context, eg *errgroup.Group, logEvent chan []byte, ns, podName, containerName string) {
	eg.Go(func() error {
		req := pkg.Cli().CoreV1().RESTClient().Get().
			Resource("pods").
			Name(podName).
			Namespace(ns).
			SubResource("log").
			VersionedParams(
				&v1.PodLogOptions{
					Container: containerName,
					Follow:    true,
				},
				scheme.ParameterCodec,
			)
		readCloser, err := req.Stream()
		if err != nil {
			fmt.Println("Stream", err)
			return err
		}
		for {
			reader := bufio.NewReader(readCloser)
			bytes, err := reader.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			logEvent <- bytes
		}
		return nil
	})

}
