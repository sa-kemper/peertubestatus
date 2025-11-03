package web

import (
	"embed"
	"errors"
	"flag"
	"html/template"
	"io/fs"
	"log"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/pkg/StatsIO"
	"github.com/sa-kemper/peertubestats/web/templates"
)

//go:embed css/*
var CssFileFS embed.FS
var ServeStaticHTTPHandler func(ResponseWriter http.ResponseWriter, Request *http.Request)
var Templates *template.Template

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
	"VideoNameToFilePath": StatsIO.VideoNameToFilePath,
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
	"urlEncode": func(s string) string {
		return url.QueryEscape(s)
	},
	"urlDecode": func(s string) (result string) {
		result, _ = url.QueryUnescape(s)
		return
	},
	"fromSafeSourceToURL": func(s string) template.URL {
		return template.URL(s)
	},
	"textInitials": func(input string) (out string) {
		input = strings.TrimSpace(input)
		initials := strings.Split(input, "_")
		for _, initial := range initials {
			out += string(initial[0])
		}
		return out
	},
	"structToUrlParams": func(input interface{}) string {
		var vals = url.Values{}
		buildParamsFromStruct(&vals, input)
		return vals.Encode()
	},
}

func buildParamsFromStruct(u *url.Values, input interface{}) {
	inputType := reflect.TypeOf(input)
	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)
		fieldValue := reflect.ValueOf(input).Field(i)
		filedKind := reflect.TypeOf(fieldValue.Interface()).Kind()
		fieldName := field.Tag.Get("form")
		if fieldName == "" || fieldName == "-" {
			continue
		}
		if timeValue, ok := fieldValue.Interface().(time.Time); ok {
			u.Add(fieldName, timeValue.Format("2006-01-02"))
			continue
		}
		if filedKind == reflect.Struct || filedKind == reflect.Slice || filedKind == reflect.Array {
			if reflect.TypeOf(fieldValue.Interface()).NumField() > 1 {
				buildParamsFromStruct(u, fieldValue.Interface())
			}
			continue
		}
		u.Add(fieldName, fieldValue.String())

	}
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
	stat, err = os.Stat("static")
	if err == nil && stat.IsDir() {
		ServeStaticHTTPHandler = http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP
	} else {
		ServeStaticHTTPHandler = http.StripPrefix("/static", http.FileServerFS(CssFileFS)).ServeHTTP
	}
}
