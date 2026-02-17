package mail

import (
	"embed"
	"log"
	"os"
	"text/template"
)

// PanicMail is a struct representing an email sent when the program is in distress and CANNOT finish the task
// it requires 3 parameters
// IncidentTimestamp the formatted time struct.
// ErrorMessage The error that caused the panic state.
// ErrorDetails The description of the error and the details to it occurring.
type PanicMail struct {
	IncidentTimestamp string
	ErrorMessage      string
	ErrorDetails      string
}

// InfoMail is a struct representing an email sent when the program notifies the administrator of an occurrence
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
