package main

import (
	"html/template"
	"log"
	"os"

	"github.com/sa-kemper/golangGetTextTest/web"
)

var Templates *template.Template

func init() {
	stat, err := os.Stat("templates")
	if err == nil && stat.IsDir() {
		Templates = template.Must(template.ParseGlob("TemplateOverride/*.gohtml"))
		return
	}
	Templates = template.Must(template.ParseFS(web.TemplateFilesFS, "*.gohtml"))

	if Templates == nil {
		log.Fatal("Templates not found")
	}
}
