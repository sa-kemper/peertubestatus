package Response

import (
	"errors"
	"net/http"

	"github.com/sa-kemper/peertubestats/i18n"
	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/web"
	"golang.org/x/text/language"
)

func (u *Utility) ReplyTemplate(writer http.ResponseWriter, request *http.Request, templateName string) {
	AcceptLanguage := request.Header.Get("Accept-Language")
	AcceptLanguage, err := ParseLanguage(AcceptLanguage)
	LogHelp.LogOnError("cannot find suitable language", map[string]string{"Accept-Language": AcceptLanguage}, err)

	templateFunctionsCopy := maps.Clone(web.TemplateFunctions)
	templateFunctionsCopy["translate"] = func(text string) string {
		lang := i18n.Languages[AcceptLanguage]
		return lang.Get(text)
	}

	TranslatedTemplate, err := u.Template.Clone()
	LogHelp.LogOnError("Copying template failed", nil, err)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	TranslatedTemplate.Funcs(templateFunctionsCopy)

	err = TranslatedTemplate.ExecuteTemplate(writer, templateName, nil)
	LogHelp.LogOnError("Failed to execute template", map[string]string{"TemplateName": templateName}, err)

}

func (u *Utility) ReplyTemplateWithData(writer http.ResponseWriter, request *http.Request, templateName string, Data interface{}) {
	writer.Header().Add("Connection", "keep-alive")
	AcceptLanguage := request.Header.Get("Accept-Language")
	tag, _, err := language.ParseAcceptLanguage(AcceptLanguage)
	LogHelp.LogOnError("Parsing Accept-Language Http Header failed", map[string]string{"Accept-Language": AcceptLanguage}, err)
	for _, langTag := range tag {
		_, ok := i18n.Languages[langTag.String()]
		if ok {
			AcceptLanguage = langTag.String()
			err = nil
			break
		}
		err = errors.New("could not find a suitable language")
	}

	if err != nil {
		AcceptLanguage = "en"
		err = nil
	}
	templateFunctionsCopy := maps.Clone(web.TemplateFunctions)
	templateFunctionsCopy["translate"] = func(text string) string {
		lang := i18n.Languages[AcceptLanguage]
		return lang.Get(text)
	}

	TranslatedTemplate := u.Template.Funcs(templateFunctionsCopy)
	err = TranslatedTemplate.ExecuteTemplate(writer, templateName, Data)
	LogHelp.LogOnError("Failed to execute template", map[string]string{"TemplateName": templateName}, err)
}
