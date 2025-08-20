package main

import (
	"embed"
	"html/template"
	"log"
	"os"
)

//go:embed web
var templatesFS embed.FS
var Templates *template.Template

func init() {
	stat, err := os.Stat("templates")
	if err == nil && stat.IsDir() {
		Templates = template.Must(template.ParseGlob("TemplateOverride/*.gohtml"))
		return
	}
	Templates = template.Must(template.ParseFS(templatesFS, "web/*.gohtml"))

	if Templates == nil {
		log.Fatal("Templates not found")
	}
}
