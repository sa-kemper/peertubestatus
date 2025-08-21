package main

import (
	"net/http"

	"github.com/sa-kemper/golangGetTextTest/web"
)

var routingTable = map[string]func(http.ResponseWriter, *http.Request){
	"/":        indexResponse,
	"/static/": http.StripPrefix("/static", http.FileServerFS(web.CssFileFS)).ServeHTTP,
}

func indexResponse(writer http.ResponseWriter, request *http.Request) {
	HandleUtility.ReplyTemplate(writer, request, "index")
}
