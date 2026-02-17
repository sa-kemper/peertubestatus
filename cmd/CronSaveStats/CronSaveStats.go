package main

import (
	"flag"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/internal/MailLog"
	"github.com/sa-kemper/peertubestats/internal/Response"
	"github.com/sa-kemper/peertubestats/pkg/StatsIO"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
)

var apiConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Username     string `json:"username"`
	Password     string `json:"-"`
	Host         string `json:"host"`
	Protocol     string `json:"protocol"`
}

// TestMail specifies if the program should just test the mail sending process and quit
var TestMail bool

func init() {
	flag.StringVar(&apiConfig.ClientId, "api-client-id", "exampleID", "Client ID")
	flag.StringVar(&apiConfig.ClientSecret, "api-client-secret", "exampleSecret", "Client Secret")
	flag.StringVar(&apiConfig.Username, "api-username", "exampleUser", "Username to authenticate with")
	flag.StringVar(&apiConfig.Password, "api-password", "examplePassword", "Password to authenticate with")
	flag.StringVar(&apiConfig.Host, "api-host", "peertube.example.com", "Host to authenticate with")
	flag.StringVar(&apiConfig.Protocol, "api-protocol", "https://", "Protocol to authenticate with")
	flag.BoolVar(&TestMail, "test-mail", false, "Test mail")
}

func main() {
	var err error
	LogHelp.AlwaysQueue = true

	err = Response.ParseConfigFromEnvFile()
	LogHelp.LogOnError("cannot parse configuration from env file", map[string]interface{}{"config": apiConfig, "smtpConfig": MailLog.SmtpConf}, err)
	LogHelp.NewLog(LogHelp.Debug, "after parsing env file the config has been changed to", map[string]interface{}{"config": apiConfig}).Log()

	err = Response.ParseConfigFromEnvironment()
	LogHelp.LogOnError("cannot parse configuration from environment", map[string]interface{}{"config": apiConfig}, err)
	LogHelp.NewLog(LogHelp.Debug, "after parsing environment variables the config has been changed to", map[string]interface{}{"config": apiConfig, "smtpConfig": MailLog.SmtpConf}).Log()

	flag.Parse()
	LogHelp.NewLog(LogHelp.Debug, "after parsing the program arguments the config has been changed to", map[string]interface{}{"config": apiConfig, "smtpConfig": MailLog.SmtpConf}).Log()

	go MailLog.SendMailOnFatalLog()

	if TestMail {
		LogHelp.NewLog(LogHelp.Debug, "Test debug message", map[string]interface{}{"config": apiConfig, "smtpConfig": MailLog.SmtpConf})
		LogHelp.NewLog(LogHelp.Info, "Test Info message", map[string]interface{}{"config": apiConfig, "smtpConfig": MailLog.SmtpConf})
		LogHelp.NewLog(LogHelp.Warn, "Test Warning message", map[string]interface{}{"config": apiConfig, "smtpConfig": MailLog.SmtpConf})
		LogHelp.NewLog(LogHelp.Error, "Test Error message", map[string]interface{}{"config": apiConfig, "smtpConfig": MailLog.SmtpConf})
		LogHelp.NewLog(LogHelp.Fatal, "Test Fatal message", map[string]interface{}{"config": apiConfig, "smtpConfig": MailLog.SmtpConf})
		select {}
	}

	PeertubeApiClient, err := peertubeApi.NewApiClient(apiConfig.ClientId, apiConfig.ClientSecret, apiConfig.Username, apiConfig.Password, apiConfig.Host, apiConfig.Protocol, peertubeApi.DEFAULT_RATE_LIMITS, nil)
	if err != nil {
		LogHelp.NewLog(LogHelp.Fatal, "error occurred during API Initialisation", map[string]interface{}{"error": err.Error()})
		println("error occurred during initialization of API client")
		panic(err)
	}
	var RawResponses [][]byte
	var collectionTime = time.Now()
	RawResponses, err = PeertubeApiClient.ListAllVideosRaw(peertubeApi.ListVideosParams{
		Count:        100,
		IsLocal:      true,
		Include:      peertubeApi.CombineVideoIncludeFlags(0, 1, 2, 4, 8, 16, 32),
		PrivacyOneOf: []int{2, 3, 4, 5},
	})
	if err != nil {
		LogHelp.NewLog(LogHelp.Fatal, "error occurred during getting video list", map[string]interface{}{"error": err.Error()})
		println("error occurred during listing of videos")
		panic(err)
	}

	serverConfig, err := PeertubeApiClient.Config()
	if err != nil {
		LogHelp.NewLog(LogHelp.Fatal, "error occurred during getting server config", map[string]interface{}{"error": err.Error()})
		println("error occurred during getting server config")
		panic(err)
	}
	StatsIO.Database.Init(PeertubeApiClient)
	err = StatsIO.Database.ImportFromRaw(RawResponses, serverConfig.ServerVersion, collectionTime)

	if err != nil {
		LogHelp.NewLog(LogHelp.Fatal, "error occurred during stats import", map[string]interface{}{"error": err.Error()}).Log()
		panic(err)
	}
}
