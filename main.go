package main

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/weibaohui/podInteractive/pkg/pod"
	"log"
	"net/http"
)

func main() {
	container := restful.NewContainer()
	ws := new(restful.WebService)
	ws.Route(ws.GET("/ns/{ns}/podName/{podName}/log").To(pod.GetContainerLog))
	ws.Route(ws.GET("/pod/").To(pod.Home))
	container.Add(ws)
	fmt.Println("SERVER 9999")

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"X-My-Header"},
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST"},
		CookiesAllowed: false,
		Container:      container}
	container.Filter(cors.Filter)

	log.Fatal(http.ListenAndServe(":9999", container))
}
