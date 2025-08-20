package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/css
var staticFS embed.FS
var staticFolder, _ = fs.Sub(staticFS, "assets")

var routingTable = map[string]func(http.ResponseWriter, *http.Request){
	"/":        indexResponse,
	"/static/": http.StripPrefix("/static", http.FileServerFS(staticFolder)).ServeHTTP,
}

func indexResponse(writer http.ResponseWriter, request *http.Request) {
	HandleUtility.ReplyTemplate(writer, request, "HelloError")
}
