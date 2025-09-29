package peertubeApi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

func (api *ApiClient) Config() (result ConfigResponse, err error) {
	const endpoint = "config"
	var response *http.Response
	response, err = api.doRequest(&http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: api.Protocol,
			Host:   api.Host,
			Path:   apiPrefix + endpoint,
		},
	})
	if err != nil {
		return result, err
	}

	err = json.NewDecoder(response.Body).Decode(&result)
	return result, err
}

// ConfigResponse represents the root structure of the JSON response
type ConfigResponse struct {
	Client             Client             `json:"client"`
	Defaults           Defaults           `json:"defaults"`
	Webadmin           Webadmin           `json:"webadmin"`
	Instance           Instance           `json:"instance"`
	Search             Search             `json:"search"`
	Plugin             Plugin             `json:"plugin"`
	Theme              Theme              `json:"theme"`
	Email              Email              `json:"email"`
	ContactForm        ContactForm        `json:"contactForm"`
	ServerVersion      string             `json:"serverVersion"`
	ServerCommit       string             `json:"serverCommit"`
	Transcoding        Transcoding        `json:"transcoding"`
	Live               Live               `json:"live"`
	VideoStudio        VideoStudio        `json:"videoStudio"`
	VideoFile          VideoFile          `json:"videoFile"`
	VideoTranscription VideoTranscription `json:"videoTranscription"`
	Import             Import             `json:"import"`
	Export             Export             `json:"export"`
	AutoBlacklist      AutoBlacklist      `json:"autoBlacklist"`
	Avatar             Avatar             `json:"avatar"`
	Banner             Banner             `json:"banner"`
	Video              Video              `json:"video"`
	VideoCaption       VideoCaption       `json:"videoCaption"`
	User               User               `json:"user"`
	VideoChannels      VideoChannels      `json:"videoChannels"`
	Trending           Trending           `json:"trending"`
	Tracker            Tracker            `json:"tracker"`
	Followings         Followings         `json:"followings"`
	Federation         Federation         `json:"federation"`
	BroadcastMessage   BroadcastMessage   `json:"broadcastMessage"`
	Homepage           Homepage           `json:"homepage"`
	OpenTelemetry      OpenTelemetry      `json:"openTelemetry"`
	Views              Views              `json:"views"`
	Storyboards        Storyboards        `json:"storyboards"`
	Webrtc             Webrtc             `json:"webrtc"`
	NsfwFlagsSettings  NsfwFlagsSettings  `json:"nsfwFlagsSettings"`
	FieldsConstraints  FieldsConstraints  `json:"fieldsConstraints"`
	Signup             Signup             `json:"signup"`
}

// Client represents the client configuration
type Client struct {
	Header       Header       `json:"header"`
	Videos       Videos       `json:"videos"`
	BrowseVideos BrowseVideos `json:"browseVideos"`
	Menu         Menu         `json:"menu"`
	OpenInApp    OpenInApp    `json:"openInApp"`
}

// Header represents the header configuration
type Header struct {
	HideInstanceName bool `json:"hideInstanceName"`
}

// Videos represents the videos configuration
type Videos struct {
	Miniature       Miniature       `json:"miniature"`
	ResumableUpload ResumableUpload `json:"resumableUpload"`
}

// Miniature represents the miniature configuration
type Miniature struct {
	PreferAuthorDisplayName bool `json:"preferAuthorDisplayName"`
}

// ResumableUpload represents the resumable upload configuration
type ResumableUpload struct {
	MaxChunkSize int `json:"maxChunkSize"`
}

// BrowseVideos represents the browse videos configuration
type BrowseVideos struct {
	DefaultSort  string `json:"defaultSort"`
	DefaultScope string `json:"defaultScope"`
}

// Menu represents the menu configuration
type Menu struct {
	Login Login `json:"login"`
}

// Login represents the login configuration
type Login struct {
	RedirectOnSingleExternalAuth bool `json:"redirectOnSingleExternalAuth"`
}

// OpenInApp represents the open in-app configuration
type OpenInApp struct {
	Android AppPlatform `json:"android"`
	Ios     AppPlatform `json:"ios"`
}

// AppPlatform represents the configuration for Android or iOS app
type AppPlatform struct {
	Intent Intent `json:"intent"`
}

// Intent represents the intent configuration for app platforms
type Intent struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Scheme      string `json:"scheme"`
	FallbackUrl string `json:"fallbackUrl"`
}

// Defaults represents the default settings
type Defaults struct {
	Publish Publish `json:"publish"`
	P2P     P2P     `json:"p2p"`
	Player  Player  `json:"player"`
}

// Publish represents the publishing settings
type Publish struct {
	DownloadEnabled bool `json:"downloadEnabled"`
	CommentsPolicy  int  `json:"commentsPolicy"`
	CommentsEnabled bool `json:"commentsEnabled"`
	Privacy         int  `json:"privacy"`
	Licence         int  `json:"licence"`
}

// P2P represents the P2P settings
type P2P struct {
	Webapp P2PSettings `json:"webapp"`
	Embed  P2PSettings `json:"embed"`
}

// P2PSettings represents the settings for webapp or embed
type P2PSettings struct {
	Enabled bool `json:"enabled"`
}

// Player represents the player settings
type Player struct {
	Theme    string `json:"theme"`
	AutoPlay bool   `json:"autoPlay"`
}

// Webadmin represents the webadmin configuration
type Webadmin struct {
	Configuration Configuration `json:"configuration"`
}

// Configuration represents the configuration settings
type Configuration struct {
	Edition Edition `json:"edition"`
}

// Edition represents the edition settings
type Edition struct {
	Allowed bool `json:"allowed"`
}

// Instance represents the instance configuration
type Instance struct {
	Name               string         `json:"name"`
	ShortDescription   string         `json:"shortDescription"`
	IsNSFW             bool           `json:"isNSFW"`
	DefaultNSFWPolicy  string         `json:"defaultNSFWPolicy"`
	DefaultClientRoute string         `json:"defaultClientRoute"`
	ServerCountry      string         `json:"serverCountry"`
	Support            Support        `json:"support"`
	Social             Social         `json:"social"`
	Customizations     Customizations `json:"customizations"`
	DefaultLanguage    string         `json:"defaultLanguage"`
	Avatars            []AvatarImage  `json:"avatars"`
	Banners            []BannerImage  `json:"banners"`
	Logo               []LogoImage    `json:"logo"`
}

// Support represents the support configuration
type Support struct {
	Text string `json:"text"`
}

// Social represents the social media links
type Social struct {
	BlueskyLink  string `json:"blueskyLink"`
	MastodonLink string `json:"mastodonLink"`
	XLink        string `json:"xLink"`
	ExternalLink string `json:"externalLink"`
}

// Customizations represents the customization settings
type Customizations struct {
	Javascript string `json:"javascript"`
	Css        string `json:"css"`
}

// AvatarImage represents an avatar image
type AvatarImage struct {
	Height    int       `json:"height"`
	Width     int       `json:"width"`
	Path      string    `json:"path"`
	FileUrl   string    `json:"fileUrl"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BannerImage represents a banner image
type BannerImage struct {
	Height    int       `json:"height"`
	Width     int       `json:"width"`
	Path      string    `json:"path"`
	FileUrl   string    `json:"fileUrl"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// LogoImage represents a logo image
type LogoImage struct {
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	Type       string `json:"type"`
	FileUrl    string `json:"fileUrl"`
	IsFallback bool   `json:"isFallback"`
}

// Search represents the search configuration
type Search struct {
	RemoteUri   RemoteUri   `json:"remoteUri"`
	SearchIndex SearchIndex `json:"searchIndex"`
}

// RemoteUri represents the remote URI settings
type RemoteUri struct {
	Users     bool `json:"users"`
	Anonymous bool `json:"anonymous"`
}

// SearchIndex represents the search index settings
type SearchIndex struct {
	Enabled            bool   `json:"enabled"`
	Url                string `json:"url"`
	DisableLocalSearch bool   `json:"disableLocalSearch"`
	IsDefaultSearch    bool   `json:"isDefaultSearch"`
}

// Plugin represents the plugin configuration
type Plugin struct {
	Registered               []RegisteredPlugin `json:"registered"`
	RegisteredExternalAuths  []interface{}      `json:"registeredExternalAuths"`
	RegisteredIdAndPassAuths []IdAndPassAuth    `json:"registeredIdAndPassAuths"`
}

// RegisteredPlugin represents a registered plugin
type RegisteredPlugin struct {
	NpmName       string                  `json:"npmName"`
	Name          string                  `json:"name"`
	Version       string                  `json:"version"`
	Description   string                  `json:"description"`
	ClientScripts map[string]ClientScript `json:"clientScripts"`
}

// ClientScript represents a client script in a plugin
type ClientScript struct {
	Script string   `json:"script"`
	Scopes []string `json:"scopes"`
}

// IdAndPassAuth represents an ID and password authentication plugin
type IdAndPassAuth struct {
	NpmName  string `json:"npmName"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	AuthName string `json:"authName"`
	Weight   int    `json:"weight"`
}

// Theme represents the theme configuration
type Theme struct {
	Registered    []RegisteredTheme `json:"registered"`
	BuiltIn       []BuiltInTheme    `json:"builtIn"`
	Default       string            `json:"default"`
	Customization Customization     `json:"customization"`
}

// RegisteredTheme represents a registered theme
type RegisteredTheme struct {
	NpmName       string                 `json:"npmName"`
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Description   string                 `json:"description"`
	Css           []string               `json:"css"`
	ClientScripts map[string]interface{} `json:"clientScripts"`
}

// BuiltInTheme represents a built-in theme
type BuiltInTheme struct {
	Name string `json:"name"`
}

// Customization represents the theme customization settings
type Customization struct {
	PrimaryColor             *string `json:"primaryColor"`
	ForegroundColor          *string `json:"foregroundColor"`
	BackgroundColor          *string `json:"backgroundColor"`
	BackgroundSecondaryColor *string `json:"backgroundSecondaryColor"`
	MenuForegroundColor      *string `json:"menuForegroundColor"`
	MenuBackgroundColor      *string `json:"menuBackgroundColor"`
	MenuBorderRadius         *string `json:"menuBorderRadius"`
	HeaderForegroundColor    *string `json:"headerForegroundColor"`
	HeaderBackgroundColor    *string `json:"headerBackgroundColor"`
	InputBorderRadius        *string `json:"inputBorderRadius"`
}

// Email represents the email configuration
type Email struct {
	Enabled bool `json:"enabled"`
}

// ContactForm represents the contact form configuration
type ContactForm struct {
	Enabled bool `json:"enabled"`
}

// Transcoding represents the transcoding configuration
type Transcoding struct {
	RemoteRunners      Enabled  `json:"remoteRunners"`
	Hls                Enabled  `json:"hls"`
	WebVideos          Enabled  `json:"web_videos"`
	EnabledResolutions []int    `json:"enabledResolutions"`
	Profile            string   `json:"profile"`
	AvailableProfiles  []string `json:"availableProfiles"`
}

// Enabled represents a generic enabled setting
type Enabled struct {
	Enabled bool `json:"enabled"`
}

// Live represents the live-streaming configuration
type Live struct {
	Enabled          bool            `json:"enabled"`
	AllowReplay      bool            `json:"allowReplay"`
	LatencySetting   Enabled         `json:"latencySetting"`
	MaxDuration      int             `json:"maxDuration"`
	MaxInstanceLives int             `json:"maxInstanceLives"`
	MaxUserLives     int             `json:"maxUserLives"`
	Transcoding      LiveTranscoding `json:"transcoding"`
	Rtmp             Rtmp            `json:"rtmp"`
}

// LiveTranscoding represents the live transcoding settings
type LiveTranscoding struct {
	Enabled            bool     `json:"enabled"`
	RemoteRunners      Enabled  `json:"remoteRunners"`
	EnabledResolutions []int    `json:"enabledResolutions"`
	Profile            string   `json:"profile"`
	AvailableProfiles  []string `json:"availableProfiles"`
}

// Rtmp represents the RTMP configuration
type Rtmp struct {
	Port int `json:"port"`
}

// VideoStudio represents the video studio configuration
type VideoStudio struct {
	Enabled       bool    `json:"enabled"`
	RemoteRunners Enabled `json:"remoteRunners"`
}

type VideoFile struct {
	Update Enabled `json:"update"`
}

// VideoTranscription represents the video transcription configuration
type VideoTranscription struct {
	Enabled       bool    `json:"enabled"`
	RemoteRunners Enabled `json:"remoteRunners"`
}

// Import represents the import configuration
type Import struct {
	Videos                      VideosImport `json:"videos"`
	VideoChannelSynchronization Enabled      `json:"videoChannelSynchronization"`
	Users                       Enabled      `json:"users"`
}

// VideosImport represents the video import settings
type VideosImport struct {
	Http    Enabled `json:"http"`
	Torrent Enabled `json:"torrent"`
}

// Export represents the export configuration
type Export struct {
	Users UsersExport `json:"users"`
}

// UsersExport represents the user export settings
type UsersExport struct {
	Enabled           bool  `json:"enabled"`
	ExportExpiration  int64 `json:"exportExpiration"`
	MaxUserVideoQuota int64 `json:"maxUserVideoQuota"`
}

// AutoBlacklist represents the auto-blacklist configuration
type AutoBlacklist struct {
	Videos VideosAutoBlacklist `json:"videos"`
}

// VideosAutoBlacklist represents the videos auto-blacklist settings
type VideosAutoBlacklist struct {
	OfUsers Enabled `json:"ofUsers"`
}

// Banner represents the banner configuration
type Banner struct {
	File File `json:"file"`
}

// File represents the file settings for avatar or banner
type File struct {
	Size       Size     `json:"size"`
	Extensions []string `json:"extensions"`
}

// Size represents the size settings
type Size struct {
	Max int64 `json:"max"`
}

// Video represents the video configuration
type Video struct {
	Image Image `json:"image"`
	File  File  `json:"file"`
}

// Image represents the image settings for videos
type Image struct {
	Extensions []string `json:"extensions"`
	Size       Size     `json:"size"`
}

// VideoCaption represents the video caption configuration
type VideoCaption struct {
	File File `json:"file"`
}

// User represents the user configuration
type User struct {
	VideoQuota      int64 `json:"videoQuota"`
	VideoQuotaDaily int64 `json:"videoQuotaDaily"`
}

// VideoChannels represents the video channels configuration
type VideoChannels struct {
	MaxPerUser int `json:"maxPerUser"`
}

// Trending represents the trending configuration
type Trending struct {
	Videos TrendingVideos `json:"videos"`
}

// TrendingVideos represents the trending videos settings
type TrendingVideos struct {
	IntervalDays int        `json:"intervalDays"`
	Algorithms   Algorithms `json:"algorithms"`
}

// Algorithms represents the trending algorithms
type Algorithms struct {
	Enabled []string `json:"enabled"`
	Default string   `json:"default"`
}

// Tracker represents the tracker configuration
type Tracker struct {
	Enabled bool `json:"enabled"`
}

// Followings represents the followings configuration
type Followings struct {
	Instance InstanceFollowings `json:"instance"`
}

// InstanceFollowings represents the instance followings settings
type InstanceFollowings struct {
	AutoFollowIndex AutoFollowIndex `json:"autoFollowIndex"`
}

// AutoFollowIndex represents the auto-follow index settings
type AutoFollowIndex struct {
	IndexUrl string `json:"indexUrl"`
}

// Federation represents the federation configuration
type Federation struct {
	Enabled bool `json:"enabled"`
}

// BroadcastMessage represents the broadcast message configuration
type BroadcastMessage struct {
	Enabled     bool   `json:"enabled"`
	Message     string `json:"message"`
	Level       string `json:"level"`
	Dismissable bool   `json:"dismissable"`
}

// Homepage represents the homepage configuration
type Homepage struct {
	Enabled bool `json:"enabled"`
}

// OpenTelemetry represents the OpenTelemetry configuration
type OpenTelemetry struct {
	Metrics Metrics `json:"metrics"`
}

// Metrics represents the metrics settings
type Metrics struct {
	Enabled               bool `json:"enabled"`
	PlaybackStatsInterval int  `json:"playbackStatsInterval"`
}

// Views represents the views configuration
type Views struct {
	Videos VideosViews `json:"videos"`
}

// VideosViews represents the videos views settings
type VideosViews struct {
	WatchingInterval WatchingInterval `json:"watchingInterval"`
}

// WatchingInterval represents the watching interval settings
type WatchingInterval struct {
	Anonymous int `json:"anonymous"`
	Users     int `json:"users"`
}

// Storyboards represents the storyboards configuration
type Storyboards struct {
	Enabled       bool    `json:"enabled"`
	RemoteRunners Enabled `json:"remoteRunners"`
}

// Webrtc represents the WebRTC configuration
type Webrtc struct {
	StunServers []string `json:"stunServers"`
}

// NsfwFlagsSettings represents the NSFW flags settings
type NsfwFlagsSettings struct {
	Enabled bool `json:"enabled"`
}

// FieldsConstraints represents the fields constraints configuration
type FieldsConstraints struct {
	Users UsersConstraints `json:"users"`
}

// UsersConstraints represents the user constraints
type UsersConstraints struct {
	Password PasswordConstraints `json:"password"`
}

// PasswordConstraints represents the password constraints
type PasswordConstraints struct {
	MinLength int `json:"minLength"`
	MaxLength int `json:"maxLength"`
}

// Signup represents the signup configuration
type Signup struct {
	Allowed                   bool `json:"allowed"`
	AllowedForCurrentIP       bool `json:"allowedForCurrentIP"`
	MinimumAge                int  `json:"minimumAge"`
	RequiresApproval          bool `json:"requiresApproval"`
	RequiresEmailVerification bool `json:"requiresEmailVerification"`
}
