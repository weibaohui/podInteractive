package docker

import (
	"bufio"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"io"
)

func Log() {
	url := "http://134.44.36.120:2376/containers/93e99eea7d4b/logs?stderr=1&stdout=1&follow=1"
	response, err := httplib.Get(url).Response()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for {
		readCloser := response.Body
		reader := bufio.NewReader(readCloser)
		bytes, err := reader.ReadBytes('\n')
		fmt.Println(string(bytes))
		if err != nil && err == io.EOF {
			break
		}

	}
}
