package Response

import (
	"log"
	"net/http"

	"github.com/sa-kemper/golangGetTextTest/internal/LogHelp"
)

func (u *Utility) ReplyTemplate(writer http.ResponseWriter, request *http.Request, templateName string) {
	err := u.Template.ExecuteTemplate(writer, templateName, nil)
	if err != nil {
		log.Println(LogHelp.NewLog(LogHelp.Error, "Failed to execute template", struct {
			TemplateName string `json:"template_name"`
			Err          string `json:"error"`
		}{templateName, err.Error()}))
	}
}
