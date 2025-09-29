package i18n

import (
	"embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/sa-kemper/peertubestats/internal/LogHelp"
)

//go:embed locales/*.mo
var localesFS embed.FS
var Languages map[string]*gotext.Mo

func init() {
	stat, err := os.Stat("locales")
	if err == nil && stat.IsDir() {
		_ = filepath.Walk("locales", func(path string, info os.FileInfo, err error) error {
			mo := gotext.NewMo()
			mo.ParseFile(path)
			Languages[mo.Language] = mo
			return nil
		})
		return
	}

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
	}
	LogHelp.NewLog(LogHelp.Info, "locale files loaded successfully", nil).Log()
}
