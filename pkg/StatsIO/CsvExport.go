package StatsIO

import (
	"flag"
	"maps"
	"strconv"

	"github.com/sa-kemper/peertubestats/i18n"
	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
	"github.com/sa-kemper/peertubestats/web/templates"
)

type CsvGenerateParameters struct {
	Videos          []peertubeApi.VideoData
	DisplaySettings templates.FrontPageRequest
	TargetLang      string
	Scope           struct {
		Views bool
		Likes bool
	}
}

func CsvGenerate(parameters CsvGenerateParameters) (csvData [][]string) {
	csvData = make([][]string, len(parameters.Videos)+1)

	var targetLang string
	if outputFlag := flag.Lookup("output-language"); outputFlag != nil {
		targetLang = outputFlag.Value.String()
	}
	if parameters.TargetLang != "" {
		targetLang = parameters.TargetLang
	}
	var Mo, found = i18n.Languages[targetLang]
	var Translate = func(id string, vars ...interface{}) string { return id }

	if !found {
		LogHelp.NewLog(LogHelp.Warn, "cannot find requested language", map[string]interface{}{"availableLanguages": maps.Keys(i18n.Languages), "requestedLang": flag.Lookup("output-language").Value.String()}).Log()
	} else {
		Translate = Mo.Get
	}

	csvData[0] = []string{Translate("Video Name"), Translate("Video URL")}
	for iterator, vid := range parameters.Videos {
		iterator++
		stats, err := ExportStats(vid.ID, parameters.DisplaySettings.Dates, parameters.DisplaySettings.Timeframe)
		var statStringSlice []string
		for _, stat := range stats {
			statStringSlice = append(statStringSlice, strconv.Itoa(int(stat.Views.Data)))
		}
		if err != nil {
			LogHelp.NewLog(LogHelp.Fatal, "cannot read stats for video", map[string]string{"error": err.Error()}).Log()
		}

		if iterator == 2 {
			// complete header
			for _, stat := range stats {
				csvData[0] = append(csvData[0], stat.Time.Format("2006-01-02"))
			}
		}

		csvData[iterator] = []string{
			vid.Name,
			"https://" + flag.Lookup("api-host").Value.String() + "/w/" + vid.ShortUUID,
		}
		// insert the stats data
		csvData[iterator] = append(csvData[iterator], statStringSlice...)
		csvData[iterator] = append(csvData[iterator], vid.Name)

	}
	csvData[0] = append(csvData[0], Translate("Video Name"))
	return csvData
}
