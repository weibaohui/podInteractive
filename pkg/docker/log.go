package docker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
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
		ContainerId: req.PathParameter("containerId"),
	}

	cancelCtx, cancel := context.WithCancel(req.Request.Context())
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
	ReadLog(ctx, readerGroup, logEvent, t)

	go func() {
		err := readerGroup.Wait()
		if err != nil {
			fmt.Println(err.Error())
		}
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
			if err := t.conn.WriteMessage(websocket.TextMessage, item); err != nil {
				cancel()
			}
		}
	}

}

func ReadLog(ctx context.Context, eg *errgroup.Group, bytes chan []byte, t *terminal) {
	eg.Go(func() error {
		url := fmt.Sprintf("http://%s/containers/%s/logs?stderr=1&stdout=1&follow=1", t.Address, t.ContainerId)
		response, err := httplib.Get(url).Response()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		reader := bufio.NewReader(response.Body)
		for {
			line, _, err := reader.ReadLine()
			bytes <- line
			if err == io.EOF {
				break
			}
		}
		return nil
	})

}
