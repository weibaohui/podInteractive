package page

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"html/template"
)

func Log(request *restful.Request, response *restful.Response) {
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
