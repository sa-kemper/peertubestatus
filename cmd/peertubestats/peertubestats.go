package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/internal/Response"
	"github.com/sa-kemper/peertubestats/pkg/StatsIO"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
	"github.com/sa-kemper/peertubestats/web"
)

type configType struct {
	BindAddress                     string
	HttpPort                        int
	MaxRequestSize                  int
	RequestTimeoutSeconds           int
	MaxConcurrentRequestConnections int
}

// config is the struct containing the necessary options used to run this program
// config behaviour: .env file < environment variables < runtime flags
var config = configType{}

var HandleUtility *Response.Utility

var apiConfig struct {
	ClientId     string
	ClientSecret string
	Username     string
	Password     string
	Host         string
	Protocol     string
}

func init() {
	flag.StringVar(&config.BindAddress, "bind-address", "127.0.0.1", "Bind address")
	flag.IntVar(&config.HttpPort, "http-port", 8080, "HTTP port")
	flag.IntVar(&config.MaxRequestSize, "max-request-size", 1048576, "Max request size")
	flag.IntVar(&config.RequestTimeoutSeconds, "request-timeout", -1, "Request timeout in seconds")
	flag.IntVar(&config.MaxConcurrentRequestConnections, "max-concurrent-request-connections", 10, "Max concurrent request connections")

	flag.StringVar(&apiConfig.ClientId, "api-client-id", "exampleID", "Client ID")
	flag.StringVar(&apiConfig.ClientSecret, "api-client-secret", "exampleSecret", "Client Secret")
	flag.StringVar(&apiConfig.Username, "api-username", "exampleUser", "Username to authenticate with")
	flag.StringVar(&apiConfig.Password, "api-password", "examplePassword", "Password to authenticate with")
	flag.StringVar(&apiConfig.Host, "api-host", "peertube.example.com", "Host to authenticate with")
	flag.StringVar(&apiConfig.Protocol, "api-protocol", "https", "Protocol to authenticate with")

}

func main() {
	var err error
	err = Response.ParseConfigFromEnvFile()
	LogHelp.LogOnError("cannot parse configuration from env file", map[string]interface{}{"config": config}, err)
	LogHelp.NewLog(LogHelp.Debug, "after parsing env file the config has been changed to", map[string]interface{}{"config": config})

	err = Response.ParseConfigFromEnvironment()
	LogHelp.LogOnError("cannot parse configuration from environment", map[string]interface{}{"config": config}, err)
	LogHelp.NewLog(LogHelp.Debug, "after parsing environment variables the config has been changed to", map[string]interface{}{"config": config})

	flag.Parse()
	LogHelp.NewLog(LogHelp.Debug, "after parsing the program arguments the config has been changed to", map[string]interface{}{"config": config})

	StatsIO.Database.Init()
	StatsIO.Database.Api, err = peertubeApi.NewApiClient(apiConfig.ClientId, apiConfig.ClientSecret, apiConfig.Username, apiConfig.Password, apiConfig.Host, apiConfig.Protocol, peertubeApi.DEFAULT_RATE_LIMITS, nil)
	if err != nil {
		println("error occurred during initialization of API client")
		panic(err)
	}

	http1Server := SetupHttpServer()

	if LogHelp.PrintableLogLevel >= 3 {
		println(`
/$$$$$$$                                 /$$               /$$                                   /$$                 /$$             
| $$__  $$                               | $$              | $$                                  | $$                | $$             
| $$  \ $$ /$$$$$$   /$$$$$$   /$$$$$$  /$$$$$$   /$$   /$$| $$$$$$$   /$$$$$$         /$$$$$$$ /$$$$$$    /$$$$$$  /$$$$$$   /$$$$$$$
| $$$$$$$//$$__  $$ /$$__  $$ /$$__  $$|_  $$_/  | $$  | $$| $$__  $$ /$$__  $$       /$$_____/|_  $$_/   |____  $$|_  $$_/  /$$_____/
| $$____/| $$$$$$$$| $$$$$$$$| $$  \__/  | $$    | $$  | $$| $$  \ $$| $$$$$$$$      |  $$$$$$   | $$      /$$$$$$$  | $$   |  $$$$$$ 
| $$     | $$_____/| $$_____/| $$        | $$ /$$| $$  | $$| $$  | $$| $$_____/       \____  $$  | $$ /$$ /$$__  $$  | $$ /$$\____  $$
| $$     |  $$$$$$$|  $$$$$$$| $$        |  $$$$/|  $$$$$$/| $$$$$$$/|  $$$$$$$       /$$$$$$$/  |  $$$$/|  $$$$$$$  |  $$$$//$$$$$$$/
|__/      \_______/ \_______/|__/         \___/   \______/ |_______/  \_______/      |_______/    \___/   \_______/   \___/ |_______/ 
                                                                                                                                      
                                                                                                                                      
                                                                                                                                      
           /$$                                                                                                                        
          | $$                                                                                                                        
  /$$$$$$$| $$   /$$                      /$$$$$$   /$$$$$$$  /$$$$$$$                                                                
 /$$_____/| $$  /$$/       /$$$$$$       /$$__  $$ /$$_____/ /$$_____/                                                                
|  $$$$$$ | $$$$$$/       |______/      | $$  \ $$|  $$$$$$ | $$                                                                      
 \____  $$| $$_  $$                     | $$  | $$ \____  $$| $$                                                                      
 /$$$$$$$/| $$ \  $$                    |  $$$$$$/ /$$$$$$$/|  $$$$$$$                                                                
|_______/ |__/  \__/                     \______/ |_______/  \_______/
`)
		print("\nListening on http://" + config.BindAddress + ":" + strconv.Itoa(config.HttpPort) + "\n")
	}

	s := *http1Server
	err = s.ListenAndServe()
	LogHelp.LogOnError("Cannot accept https connection", nil, err)
	select {}
}

func SetupHttpServer() (server *http.Server) {
	allowedProtocols := new(http.Protocols)
	allowedProtocols.SetHTTP1(true)
	allowedProtocols.SetHTTP2(true)

	http2Configuration := new(http.HTTP2Config)
	http2Configuration.PingTimeout = time.Duration(config.RequestTimeoutSeconds) * time.Second
	http2Configuration.MaxConcurrentStreams = config.MaxConcurrentRequestConnections
	http2Configuration.MaxReceiveBufferPerStream = config.MaxRequestSize

	HandleUtility = &Response.Utility{
		Template: web.Templates,
	}

	serveMux := http.NewServeMux()
	for path, function := range routingTable {
		serveMux.HandleFunc(path, function)
	}

	http1Server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", config.BindAddress, config.HttpPort),
		Handler:           serveMux,
		Protocols:         allowedProtocols,
		ReadHeaderTimeout: time.Duration(config.RequestTimeoutSeconds) * time.Second,
		ReadTimeout:       time.Duration(config.RequestTimeoutSeconds) * time.Second,
		WriteTimeout:      time.Duration(config.RequestTimeoutSeconds) * time.Second,

		BaseContext: func(net net.Listener) context.Context {
			return context.WithValue(context.Background(), Response.UtilityIndex, HandleUtility)
		},
	}

	return http1Server
}
