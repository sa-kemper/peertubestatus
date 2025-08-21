package i18n

import (
	"embed"
	"fmt"
	"strings"

	"github.com/leonelquinteros/gotext"
)

//go:embed locales/*.mo
var localesFS embed.FS
var Languages map[string]*gotext.Mo

func init() {
	files, err := localesFS.ReadDir("locales")
	if err != nil {
		panic(err)
	}
	Languages = make(map[string]*gotext.Mo)
	for _, file := range files {
		fileName := file.Name()
		fileName = strings.TrimSuffix(fileName, ".mo")
		lang := gotext.NewMoFS(localesFS)
		lang.ParseFile("locales/" + fileName + ".mo")
		Languages[fileName] = lang

		fmt.Printf("fileName=%s\n", fileName)
	}
	fmt.Println("Langcode en, de:", Languages["en"].Get("languagecode"), Languages["de"].Get("languagecode"))
}
