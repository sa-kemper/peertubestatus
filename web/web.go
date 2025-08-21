package web

import (
	"embed"
	"io/fs"
)

//go:embed css/*
var CssFileFS embed.FS

//var CssFileFS, _ = fs.Sub(cssFs, "css")

//go:embed templates/*.gohtml
var templatesFS embed.FS
var TemplateFilesFS, err = fs.Sub(templatesFS, "templates")
