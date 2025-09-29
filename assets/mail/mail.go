package mail

import (
	"embed"
	"log"
	"os"
	"text/template"
)

type PanicMail struct {
	IncidentTimestamp int
	ErrorMessage      string
	ErrorDetails      string
}

type InfoMail struct {
	IncidentTimestamp string
	Message           string
}

type WarningMail struct {
	IncidentTimestamp string
	WarningMessage    string
	AdditionalDetails string
}

//go:embed *.tmpl
var templatesFS embed.FS

var TemplateFunctions = template.FuncMap{}
var Templates *template.Template

func init() {
	stat, err := os.Stat("templates")
	if err == nil && stat.IsDir() {
		Templates = template.New("TestTemplate")
		Templates.Funcs(TemplateFunctions)
		Templates, err = Templates.ParseGlob("TemplateOverride/*.tmpl")
		template.Must(Templates, err)

		return
	}
	Templates = template.New("TestTemplate")
	Templates = Templates.Funcs(TemplateFunctions)
	Templates, err = Templates.ParseFS(templatesFS, "*.tmpl")
	template.Must(Templates, err)

	if Templates == nil {
		log.Fatal("Templates not found")
	}
}
