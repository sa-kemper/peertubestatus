package web

import (
	"embed"
	"errors"
	"flag"
	"html/template"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/pkg/StatsIO"
	"github.com/sa-kemper/peertubestats/web/templates"
)

//go:embed css/*
var CssFileFS embed.FS

//var CssFileFS, _ = fs.Sub(cssFs, "css")

//go:embed templates/*.gohtml
var templatesFS embed.FS
var TemplateFilesFS, _ = fs.Sub(templatesFS, "templates")
var TemplateFunctions = template.FuncMap{
	"translate": func(text string) string {
		return "" // This is a placeholder function that will be replaced based on the current http request.
		// Don't worry, the original translate function is not touched, we copy the FuncMap, replace one function and then provide it to the templating engine, this ensures no race conditions.
	},
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		// All credit to 82219/tux21b from stackoverflow.
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
	"videoStats": func(videoID int64, request templates.FrontPageRequest) (stats []StatsIO.VideoStat) {
		var err error
		stats, err = StatsIO.ExportStats(videoID, request.Dates, request.Timeframe)
		LogHelp.LogOnError("Cannot retrieve stats", map[string]interface{}{"videoID": videoID, "request": request}, err)
		return stats
	},
	"formatDate": func(date time.Time) string {
		return date.Format("2006-01-02")
	},
	"formatDuration": func(date time.Duration) string {
		return date.String()
	},
	"flagGet": func(key string) string {
		flag := flag.Lookup(key)
		if flag == nil {
			return ""
		}
		return flag.Value.String()
	},
}

func init() {
	stat, err := os.Stat("templates")
	if err == nil && stat.IsDir() {
		Templates = template.New("TestTemplate")
		Templates.Funcs(TemplateFunctions)
		Templates, err = Templates.ParseGlob("TemplateOverride/*.gohtml")
		template.Must(Templates, err)

		return
	}
	Templates = template.New("TestTemplate")
	Templates = Templates.Funcs(TemplateFunctions)
	Templates, err = Templates.ParseFS(TemplateFilesFS, "*.gohtml")
	template.Must(Templates, err)

	if Templates == nil {
		log.Fatal("Templates not found")
	}
}

var Templates *template.Template
