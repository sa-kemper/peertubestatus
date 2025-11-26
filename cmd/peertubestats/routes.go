package main

import (
	"encoding/csv"
	"errors"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sa-kemper/peertubestats/i18n"
	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/internal/Response"
	"github.com/sa-kemper/peertubestats/pkg/StatsIO"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
	"github.com/sa-kemper/peertubestats/web"
	"github.com/sa-kemper/peertubestats/web/templates"
	"golang.org/x/text/language"
)

var routingTable = map[string]func(http.ResponseWriter, *http.Request){
	"/":                        referToIndex,
	"/Video":                   VideoIndex,
	"/static/":                 web.ServeStaticHTTPHandler,
	"/Video/{id}":              singleVideoPage,
	"/Video/csv":               csvDownload,
	"/lazy-static/thumbnails/": http.FileServer(http.Dir(path.Join(StatsIO.Database.DataFolder, ""))).ServeHTTP,
}

func referToIndex(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Location", "/Video")
	writer.WriteHeader(http.StatusPermanentRedirect)
}

func csvDownload(writer http.ResponseWriter, request *http.Request) {
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

	videos, err := StatsIO.GetAllVideos()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		LogHelp.NewLog(LogHelp.Error, "cannot obtain videos", map[string]string{"error": err.Error()}).Log()
		return
	}
	var requestParameters templates.FrontPageRequest
	_ = Response.BindToStruct(request, &requestParameters)
	data := StatsIO.CsvGenerate(StatsIO.CsvGenerateParameters{
		Videos:          videos,
		DisplaySettings: requestParameters,
		TargetLang:      AcceptLanguage,
		Scope: struct {
			Views bool
			Likes bool
		}{Views: true},
	})

	writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
	writer.Header().Set("Content-Disposition", "attachment; filename=\"stats-from"+time.Now().Format("2006-01-02")+".csv\"")
	writer.WriteHeader(http.StatusOK)
	csvWriter := csv.NewWriter(writer)
	err = csvWriter.WriteAll(data)
	LogHelp.LogOnError("cannot write csv data", nil, err)
}

func singleVideoPage(writer http.ResponseWriter, request *http.Request) {
	util := request.Context().Value(Response.UtilityIndex)
	utility := util.(*Response.Utility)

	videoParam := request.PathValue("id")
	videoId, err := strconv.Atoi(videoParam)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	video, err := StatsIO.GetVideo(int64(videoId))
	LogHelp.LogOnError("cannot obtain video", map[string]interface{}{"videoID": videoId}, err)

	var FrontPageForm templates.FrontPageRequest
	err = Response.BindToStruct(request, &FrontPageForm)
	LogHelp.LogOnError("cannot bind front page", map[string]interface{}{"videoID": videoId, "request": request}, err)
	FrontPageForm.HandleZeroDate()

	utility.ReplyTemplateWithData(writer, request, "singleVideo", struct {
		Video   peertubeApi.VideoData
		Request templates.FrontPageRequest
	}{
		Request: FrontPageForm,
		Video:   video,
	})
	request.Close = true
}

func VideoIndex(writer http.ResponseWriter, request *http.Request) {
	util := request.Context().Value(Response.UtilityIndex)
	utility := util.(*Response.Utility)

	var AllVideos, err = StatsIO.GetAllVideos()
	var Videos []peertubeApi.VideoData
	if err != nil {
		LogHelp.NewLog(LogHelp.Fatal, "cannot load video database", map[string]interface{}{"error": err.Error()}).Log()
		os.Exit(2)
	}
	var FrontPageForm templates.FrontPageRequest
	err = Response.BindToStruct(request, &FrontPageForm)
	FrontPageForm.HandleZeroDate()
	LogHelp.LogOnError("cannot bind reuest to struct", map[string]interface{}{"request": request, "struct": FrontPageForm}, err)

	if FrontPageForm.Query == "" {
		Videos = AllVideos
	} else {
		FrontPageForm.Query = strings.ToLower(FrontPageForm.Query)
		for _, video := range AllVideos {
			if strings.Contains(strings.ToLower(video.Name), FrontPageForm.Query) || strings.Contains(strings.ToLower(video.Channel.Name), FrontPageForm.Query) || strings.Contains(strings.ToLower(video.Account.Name), FrontPageForm.Query) || strings.Contains(strings.ToLower(video.Channel.Name), FrontPageForm.Query) || strings.Contains(strings.ToLower(video.Category.Label), FrontPageForm.Query) || strings.Contains(strings.ToLower(video.TruncatedDescription), FrontPageForm.Query) {
				Videos = append(Videos, video)
			}
		}
	}
	var summaryBucket []StatsIO.VideoStat
	for _, video := range Videos {
		currentBucket, err := StatsIO.ExportStats(video.ID, FrontPageForm.Dates, FrontPageForm.Timeframe)
		if err != nil {
			LogHelp.LogOnError("cannot export stats", map[string]interface{}{"VideoID": video.ID}, err)
			return
		}
		if len(summaryBucket) == 0 {
			summaryBucket = make([]StatsIO.VideoStat, len(currentBucket))
		}

		for index, val := range currentBucket {
			if stat := summaryBucket[index]; stat.Time.IsZero() {
				summaryBucket[index] = val
				continue
			}
			summaryBucket[index].Views.Data += val.Views.Data
			summaryBucket[index].Likes.Data += val.Likes.Data
		}
	}

	summaryBucket = StatsIO.PrepareStatsBucketWithAverages(summaryBucket)
	TotalViews := summaryBucket[max(0, len(summaryBucket)-1)].Views.Data
	TotalLikes := summaryBucket[max(0, len(summaryBucket)-1)].Likes.Data
	utility.ReplyTemplateWithData(writer, request, "index", map[string]interface{}{"Request": FrontPageForm, "Videos": Videos, "Summary": struct {
		Chart         []StatsIO.VideoStat
		TotalViews    int64
		TotalLikes    int64
		LikeViewRatio float64
	}{Chart: summaryBucket, TotalViews: TotalViews, TotalLikes: TotalLikes, LikeViewRatio: float64(TotalLikes) / float64(max(1, TotalViews))}})
}
