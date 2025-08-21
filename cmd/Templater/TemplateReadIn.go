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
		Templates = template.New("TestTemplate")
		Templates.Funcs(template.FuncMap{
			"translate": func(text string) string {
				return ""
			},
		})
		Templates, err = Templates.ParseGlob("TemplateOverride/*.gohtml")
		template.Must(Templates, err)

		return
	}
	Templates = template.New("TestTemplate")
	Templates = Templates.Funcs(template.FuncMap{
		"translate": func(text string) string {
			return ""
		},
	})
	Templates, err = Templates.ParseFS(web.TemplateFilesFS, "*.gohtml")
	template.Must(Templates, err)

	if Templates == nil {
		log.Fatal("Templates not found")
	}
}
