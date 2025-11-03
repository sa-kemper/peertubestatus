package main

import (
	"encoding/csv"
	"flag"
	"maps"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/sa-kemper/peertubestats/i18n"
	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/internal/MailLog"
	"github.com/sa-kemper/peertubestats/internal/Response"
	"github.com/sa-kemper/peertubestats/pkg/StatsIO"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
	"github.com/sa-kemper/peertubestats/web"
	"github.com/sa-kemper/peertubestats/web/templates"
)

var Config struct {
	OutputFolder    string
	OutputLanguage  string
	StartDateParam  string
	EndDateParam    string
	SampleFrequency string
	ApiHost         string
}

func init() {
	flag.StringVar(&Config.OutputFolder, "output", "./Reports", "Output folder")
	flag.StringVar(&Config.OutputLanguage, "output-language", "de", "Output language, must have a available locales file")
	flag.StringVar(&Config.StartDateParam, "start-date", "", "Start date")
	flag.StringVar(&Config.EndDateParam, "end-date", "", "End date")
	flag.StringVar(&Config.SampleFrequency, "sample-frequency", "Daily", "Sample frequency can either be (Daily, Monthly, Yearly).")
	flag.StringVar(&Config.ApiHost, "api-host", "peertube.example.com", "peertube API host")
}

func main() {
	var err error
	err = Response.ParseConfigFromEnvFile()
	LogHelp.LogOnError("cannot parse configuration from env file", map[string]interface{}{"config": Config}, err)
	LogHelp.NewLog(LogHelp.Debug, "after parsing env file the config has been changed to", map[string]interface{}{"config": Config})

	err = Response.ParseConfigFromEnvironment()
	LogHelp.LogOnError("cannot parse configuration from environment", map[string]interface{}{"config": Config}, err)
	LogHelp.NewLog(LogHelp.Debug, "after parsing environment variables the config has been changed to", map[string]interface{}{"config": Config})

	flag.Parse()

	LogHelp.AlwaysQueue = true
	go MailLog.SendMailOnFatalLog()

	StatsIO.Database.Init(nil)
	videos, err := StatsIO.GetAllVideos()
	if err != nil {
		LogHelp.NewLog(LogHelp.Fatal, "cannot get all videos", map[string]interface{}{"errors": err, "videos": videos}).Log()
	}

	StartDate, err := time.Parse("2006.01.02", Config.StartDateParam)
	if Config.StartDateParam != "" {
		LogHelp.LogOnError("cannot parse start date", map[string]interface{}{"startDate": Config.StartDateParam}, err)
	}
	EndDate, err := time.Parse("2006.01.02", Config.EndDateParam)
	if Config.EndDateParam != "" {
		LogHelp.NewLog(LogHelp.Fatal, "cannot parse end date", map[string]string{"error": err.Error()}).Log()
	}

	var DisplaySettings = templates.FrontPageRequest{
		Timeframe: Config.SampleFrequency,
		Dates: templates.TwoDateForm{
			StartDate: StartDate,
			EndDate:   EndDate,
		},
	}
	DisplaySettings.HandleZeroDate()

	err = os.MkdirAll(Config.OutputFolder, 0700)
	LogHelp.LogOnError("cannot create output directory", map[string]interface{}{"outputFolder": Config.OutputFolder}, err)

	err = os.MkdirAll(filepath.Join(Config.OutputFolder, "static"), 0700)
	LogHelp.LogOnError("cannot create static style directory", map[string]interface{}{"outputFolder": Config.OutputFolder}, err)

	go func() {
		styleBytes, err := web.CssFileFS.ReadFile("css/style.css")
		err = os.WriteFile(filepath.Join(Config.OutputFolder, "static", "style.css"), styleBytes, 0600)
		LogHelp.LogOnError("cannot create static style directory", map[string]interface{}{"outputFolder": Config.OutputFolder}, err)
	}()

	Config.OutputLanguage, err = Response.ParseLanguage(Config.OutputLanguage)
	LogHelp.LogOnError("cannot parse language", map[string]string{"language": Config.OutputLanguage}, err)

	// while the reports are being generated, handle the writing of the views.csv
	fileHandle, localErr := os.OpenFile(filepath.Join(Config.OutputFolder, "views.csv"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if localErr != nil {
		LogHelp.NewLog(LogHelp.Fatal, "cannot create views.csv", map[string]string{"error": localErr.Error()}).Log()
	}
	defer fileHandle.Close()
	writer := csv.NewWriter(fileHandle)
	defer writer.Flush()
	localErr = writer.WriteAll(StatsIO.CsvGenerate(StatsIO.CsvGenerateParameters{
		Videos:          videos,
		DisplaySettings: DisplaySettings,
		Scope: struct {
			Views bool
			Likes bool
		}{
			Views: true,
			Likes: false,
		},
	}))
	if localErr != nil {
		LogHelp.NewLog(LogHelp.Fatal, "cannot write to views.csv", map[string]string{"error": localErr.Error()}).Log()
	}

	// while the reports are being generated, output an index page.
	fileHandler, LocalErr := os.OpenFile(filepath.Join(Config.OutputFolder, "index.html"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if LocalErr != nil {
		LogHelp.NewLog(LogHelp.Fatal, "cannot create index.html", map[string]string{"error": LocalErr.Error()}).Log()
		return
	}
	defer fileHandler.Close()

	TranslatedTemplate, LocalErr := web.Templates.Clone()
	if LocalErr != nil {
		LogHelp.NewLog(LogHelp.Fatal, "cannot clone template", map[string]string{"error": LocalErr.Error()}).Log()
	}

	translatedFunctions := maps.Clone(web.TemplateFunctions)
	translatedFunctions["translate"] = func(text string) string {
		lang := i18n.Languages[Config.OutputLanguage]
		return lang.Get(text)
	}

	LocalErr = TranslatedTemplate.Funcs(translatedFunctions).ExecuteTemplate(fileHandler, "reportIndex", map[string]interface{}{"Videos": videos})
	LogHelp.LogOnError("cannot write index report page", nil, LocalErr)

	for _, vid := range videos {
		filePath := path.Join(Config.OutputFolder, "ReportFor_"+StatsIO.VideoNameToFilePath(vid.Name)+".html")
		absFilePath, err := filepath.Abs(filePath)
		LogHelp.LogOnError("cannot find absolute file path", map[string]string{"filePath": filePath}, err)

		fHandler, err := os.OpenFile(absFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		LogHelp.LogOnError("cannot open report file", map[string]interface{}{"filename": "ReportFor_" + StatsIO.VideoNameToFilePath(vid.Name) + ".html"}, err)
		if err != nil {
			continue
		}

		err = TranslatedTemplate.ExecuteTemplate(fHandler, "singleVideoExport", struct {
			Video   peertubeApi.VideoData
			Request templates.FrontPageRequest
		}{
			Request: DisplaySettings,
			Video:   vid,
		})
		if err != nil {
			LogHelp.NewLog(LogHelp.Fatal, "cannot output report file", map[string]interface{}{
				"filename":        "ReportFor_" + StatsIO.VideoNameToFilePath(vid.Name) + ".html",
				"videoID":         vid.ID,
				"displaySettings": DisplaySettings,
				"outputFolder":    Config.OutputFolder,
				"outputLanguage":  Config.OutputLanguage,
				"startDate":       StartDate,
				"endDate":         EndDate,
				"error":           err,
			}).Log()
		}

		err = fHandler.Close()
		LogHelp.LogOnError("cannot close file", map[string]interface{}{"filename": "ReportFor_" + StatsIO.VideoNameToFilePath(vid.Name) + ".html"}, err)
	}
}
