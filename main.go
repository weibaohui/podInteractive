package main

import (
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/weibaohui/podInteractive/pkg/page"
	"github.com/weibaohui/podInteractive/pkg/pod"
	"github.com/weibaohui/podInteractive/pkg/utils"
	"log"
	"net/http"
)

func main() {
	var path string
	flag.StringVar(&path, "kubeconfig", "", "kubernetes config path")
	flag.Parse()
	utils.SetPath(path)

	fmt.Println("SERVER 9999")
	container := restful.NewContainer()
	ws := new(restful.WebService)
	ws.Route(ws.GET("/ns/{ns}/podName/{podName}/log").To(pod.PodLog))
	ws.Route(ws.GET("/ns/{ns}/podName/{podName}/exec").To(pod.PodExec))
	ws.Route(ws.GET("/log/").To(page.Log))
	ws.Route(ws.GET("/exec/").To(page.Exec))
	container.Add(ws)

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"X-My-Header"},
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST"},
		CookiesAllowed: false,
		Container:      container}
	container.Filter(cors.Filter)
	container.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	log.Fatal(http.ListenAndServe(":9999", container))
}
