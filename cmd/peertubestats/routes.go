package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/internal/Response"
	"github.com/sa-kemper/peertubestats/pkg/StatsIO"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
	"github.com/sa-kemper/peertubestats/web"
	"github.com/sa-kemper/peertubestats/web/templates"
)

var routingTable = map[string]func(http.ResponseWriter, *http.Request){
	"/":           indexResponse,
	"/static/":    http.StripPrefix("/static", http.FileServerFS(web.CssFileFS)).ServeHTTP,
	"/Video/{id}": singleVideoPage,
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
}

func indexResponse(writer http.ResponseWriter, request *http.Request) {
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
		for _, video := range AllVideos {
			if strings.Contains(video.Name, FrontPageForm.Query) {
				Videos = append(Videos, video)
			}

			if strings.Contains(video.Channel.Name, FrontPageForm.Query) {
				Videos = append(Videos, video)
			}

			if strings.Contains(video.Account.Name, FrontPageForm.Query) {
				Videos = append(Videos, video)
			}
		}
	}

	utility.ReplyTemplateWithData(writer, request, "index", map[string]interface{}{"Request": FrontPageForm, "Videos": Videos})
}
