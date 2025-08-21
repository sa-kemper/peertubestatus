package i18n

import (
	"embed"
	"fmt"
	"strings"

	"github.com/leonelquinteros/gotext"
)

//go:embed locales/*.mo
var localesFS embed.FS
var Languages map[string]*gotext.Locale

func init() {
	files, err := localesFS.ReadDir("locales")
	if err != nil {
		panic(err)
	}
	Languages = make(map[string]*gotext.Locale)
	for _, file := range files {
		fileName := file.Name()
		fileName = strings.TrimSuffix(fileName, ".mo")
		Languages[fileName] = gotext.NewLocaleFS(fileName, localesFS)
		fmt.Printf("fileName=%s\n", fileName)
	}
	fmt.Println("Langcode en, de:", Languages["en"].Get("languagecode"), Languages["de"].Get("languagecode"))
}
