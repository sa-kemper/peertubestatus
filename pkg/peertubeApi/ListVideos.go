package peertubeApi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"time"
)

func (api *ApiClient) ListVideos(args ListVideosParams) (response VideoResponse, err error) {
	const endpoint = "videos"
	_, err = ValidateVideoIncludeFlags(args.Include, api.isAdmin)
	if err != nil {
		return
	}

	var listVideosUrl = url.URL{
		Scheme:     api.Protocol,
		Host:       api.Host,
		Path:       apiPrefix + endpoint,
		ForceQuery: true,
		RawQuery:   toQueryParams(args).Encode(),
	}
	listVideosUrl.Query().Add("host", api.Host)

	httpResponse, err := api.doRequest(&http.Request{
		Method: http.MethodGet,
		URL:    &listVideosUrl,
		Header: api.headers,
		Host:   api.Host,
	})
	if err != nil {
		return VideoResponse{}, err
	}
	if httpResponse.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(httpResponse.Body)
		return VideoResponse{}, errors.New("http status: " + httpResponse.Status + "\nresponse body: " + string(responseBody))
	}

	err = json.NewDecoder(httpResponse.Body).Decode(&response)
	if err != nil {
		return VideoResponse{}, err
	}
	return response, nil
}

// ListVideosParams represents the query parameters for listing videos in the PeerTube API
type ListVideosParams struct {
	// AutoTagOneOf filters videos by automatic tags (PeerTube >= 6.2, admin/moderator only)
	// These are predefined automatic tags that can be used to filter videos
	AutoTagOneOf []string

	// VideoCategorySet filters videos by their category IDs
	// Corresponds to the categoryOneOf parameter in the API documentation
	VideoCategorySet []int

	// Count specifies the number of items to return in the response
	// Default is 15 if not specified
	Count int

	// ExcludeAlreadyWatched determines whether to exclude videos from the user's watch history
	ExcludeAlreadyWatched bool

	// HasHLSFiles filters to show only videos with HLS (HTTP Live Streaming) files
	// Requires PeerTube >= 4.0
	HasHLSFiles bool

	// HasWebVideoFiles filters to show only videos with Web Video files
	// Requires PeerTube >= 6.0
	HasWebVideoFiles bool

	// Include allows additional video details to be included in the results
	// Only usable by administrators and moderators
	// Can be combined using bitwise OR operator
	// Possible values:
	// 0 - NONE
	// 1 - NOT_PUBLISHED_STATE
	// 2 - BLACKLISTED
	// 4 - BLOCKED_OWNER
	// 8 - FILES
	// 16 - CAPTIONS
	// 32 - VIDEO SOURCE
	Include VideoIncludeFlags

	// IncludeScheduledLive determines whether to include live videos scheduled for later
	IncludeScheduledLive bool

	// IsLive filters to show only live videos
	IsLive bool

	// IsLocal filters to show only local objects (PeerTube >= 4.0)
	IsLocal bool

	// VideoLanguageSet filters videos by their language IDs
	// Use "_unknown" to filter videos without a specified language
	VideoLanguageSet []string

	// VideoLicenseSet filters videos by their license IDs
	VideoLicenseSet []string

	// Nsfw determines whether to include NSFW (Not Safe For Work) videos
	Nsfw bool

	// NsfwFlagsExcluded excludes videos with specific NSFW flags
	// Possible values: 0, 1, 2, 4
	NsfwFlagsExcluded int

	// NsfwFlagsIncluded includes videos with specific NSFW flags
	// Possible values: 0, 1, 2, 4
	NsfwFlagsIncluded int

	// PrivacyOneOf filters videos by their privacy settings (PeerTube >= 4.0)
	// Possible values: 1, 2, 3, 4, 5
	PrivacyOneOf []int

	// Sort specifies the sorting method for the video list
	// Available values:
	// - "name"
	// - "-duration"
	// - "-createdAt"
	// - "-publishedAt"
	// - "-views"
	// - "-likes"
	// - "-comments"
	// - "-trending"
	// - "-hot"
	// - "-best"
	//Sort string

	// Start is the offset used to paginate results
	Start int

	// TagsAllOf filters videos that have ALL the specified tags
	TagsAllOf []string

	// TagsOneOf filters videos that have ANY of the specified tags
	TagsOneOf []string
}

type VideoResponse struct {
	Total int64       `json:"total"`
	Data  []VideoData `json:"data"`
}

type VideoData struct {
	ID                    int64           `json:"id"`
	UUID                  string          `json:"uuid"`
	ShortUUID             string          `json:"shortUUID"`
	IsLive                bool            `json:"isLive"`
	LiveSchedules         []LiveSchedule  `json:"liveSchedules"`
	CreatedAt             string          `json:"createdAt"`
	PublishedAt           string          `json:"publishedAt"`
	UpdatedAt             string          `json:"updatedAt"`
	OriginallyPublishedAt string          `json:"originallyPublishedAt"`
	Category              Metadata        `json:"category"`
	Licence               Metadata        `json:"licence"`
	Language              Metadata        `json:"language"`
	Privacy               Metadata        `json:"privacy"`
	TruncatedDescription  string          `json:"truncatedDescription"`
	Duration              int64           `json:"duration"`
	AspectRatio           float64         `json:"aspectRatio"`
	IsLocal               bool            `json:"isLocal"`
	Name                  string          `json:"name"`
	ThumbnailPath         string          `json:"thumbnailPath"`
	PreviewPath           string          `json:"previewPath"`
	EmbedPath             string          `json:"embedPath"`
	Views                 int64           `json:"views"`
	Likes                 int64           `json:"likes"`
	Dislikes              int64           `json:"dislikes"`
	Comments              int64           `json:"comments"`
	NSFW                  bool            `json:"nsfw"`
	NSFWFlags             int64           `json:"nsfwFlags"`
	NSFWSummary           string          `json:"nsfwSummary"`
	WaitTranscoding       bool            `json:"waitTranscoding"`
	State                 Metadata        `json:"state"`
	ScheduledUpdate       ScheduledUpdate `json:"scheduledUpdate"`
	Blacklisted           bool            `json:"blacklisted"`
	BlacklistedReason     string          `json:"blacklistedReason"`
	Account               Account         `json:"account"`
	Channel               Channel         `json:"channel"`
	UserHistory           UserHistory     `json:"userHistory"`
}

func (v *VideoData) GetPublishedAt() (time.Time, error) {
	return time.Parse(time.RFC3339Nano, v.PublishedAt)
}

type LiveSchedule struct {
	StartAt string `json:"startAt"`
}

type Metadata struct {
	ID    interface{} `json:"id"`
	Label string      `json:"label"`
}

type ScheduledUpdate struct {
	Privacy  int64  `json:"privacy"`
	UpdateAt string `json:"updateAt"`
}

type Account struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	URL         string   `json:"url"`
	Host        string   `json:"host"`
	Avatars     []Avatar `json:"avatars"`
}

type Channel struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	URL         string   `json:"url"`
	Host        string   `json:"host"`
	Avatars     []Avatar `json:"avatars"`
}

type Avatar struct {
	Path      string `json:"path"`
	Width     int64  `json:"width"`
	Height    int64  `json:"height"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type UserHistory struct {
	CurrentTime int64 `json:"currentTime"`
}

func (api *ApiClient) ListAllVideos(params ListVideosParams) (videos []VideoData, err error) {
	var totalVideos int64 = 10
	var i int64
	for i = 0; i < totalVideos; {
		params.Start = int(i)
		response, err := api.ListVideos(params)
		if err != nil {
			return nil, err
		}
		videos = slices.Concat(videos, response.Data)
		totalVideos = response.Total
		i += int64(len(response.Data))

	}
	return videos, err
}
