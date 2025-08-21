package Response

import (
	"html/template"
	"log"
	"maps"
	"net/http"
	"slices"

	"github.com/sa-kemper/golangGetTextTest/i18n"
	"github.com/sa-kemper/golangGetTextTest/internal/LogHelp"
)

func (u *Utility) ReplyTemplate(writer http.ResponseWriter, request *http.Request, templateName string) {
	TranslatedTemplate := u.Template.Funcs(template.FuncMap{
		"translate": func(text string) string {
			lang, ok := i18n.Languages[request.Header.Get("Accept-Language")[0:2]]
			if !ok {
				log.Println(LogHelp.NewLog(LogHelp.Warn, "Accept-Language is not supported", struct {
					RequestedLang string   `json:"requested_lang"`
					AvailableLang []string `json:"available_lang"`
				}{
					RequestedLang: request.Header.Get("Accept-Language")[0:2],
					AvailableLang: slices.Sorted(maps.Keys(i18n.Languages)),
				}))
			}
			return lang.Get(text)
		},
	})
	err := TranslatedTemplate.ExecuteTemplate(writer, templateName, nil)
	if err != nil {
		log.Println(LogHelp.NewLog(LogHelp.Error, "Failed to execute template", struct {
			TemplateName string `json:"template_name"`
			Err          string `json:"error"`
		}{templateName, err.Error()}))
	}
}
