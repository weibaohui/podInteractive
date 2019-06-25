package docker

import (
	"bufio"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/websocket"
	"io"
	"log"
)

func Log(req *restful.Request, resp *restful.Response) {

	c, err := upgrader.Upgrade(resp, req.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	t := &terminal{
		conn:        c,
		Address:     "134.44.36.120:2376",
		ContainerId: "578445bc",
	}
	url := fmt.Sprintf("http://%s/containers/%s/logs?stderr=1&stdout=1&follow=1", t.Address, t.ContainerId)
	response, err := httplib.Get(url).Response()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	readCloser := response.Body
	reader := bufio.NewReader(readCloser)
	for {
		line, _, err := reader.ReadLine()
		c.WriteMessage(websocket.TextMessage, line)
		if err == io.EOF {
			break
		}

	}
}
