package main

import (
	"flag"
	"strconv"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/pkg/StatsIO"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
	"github.com/sa-kemper/peertubestats/web/templates"
)

func csvGenerate(videos []peertubeApi.VideoData, displaySettings templates.FrontPageRequest) (csvData [][]string) {
	//var Translate = i18n.Languages[Config.OutputLanguage].Get // TODO: Maybe add translation for this.
	csvData = make([][]string, len(videos)+1)
	csvData[0] = []string{"Video Name", "Video URL"}
	for iterator, vid := range videos {
		iterator++
		stats, err := StatsIO.ExportStats(vid.ID, displaySettings.Dates, displaySettings.Timeframe)
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

	}
	return csvData
}
