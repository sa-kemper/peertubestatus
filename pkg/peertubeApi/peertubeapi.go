package peertubeApi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
)

// DEFAULT_RATE_LIMITS implements the default rate limits provided by https://docs.joinpeertube.org/api-rest-reference.html
var DEFAULT_RATE_LIMITS = map[endpointPath]*RateLimit{
	endpointPath("/"):                            NewRateLimit("/", 50, time.Second*10),
	endpointPath("/users/token"):                 NewRateLimit("/users/token", 15, time.Minute*5),
	endpointPath("/users/register"):              NewRateLimit("/users/register", 2, time.Minute*5),
	endpointPath("/users/ask-send-verify-email"): NewRateLimit("/users/ask-send-verify-email", 15, time.Minute*5),
}

type tokenLoginData struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type ApiClient struct {
	clientId     string
	clientSecret string
	doRequest    func(req *http.Request) (*http.Response, error)
	accessToken  string
	Host         string
	Protocol     string
	RateLimit    RateLimitMap
	headers      http.Header
	tokenData    *tokenLoginData
	isAdmin      bool
}

func (api *ApiClient) ListAllVideosRaw(params ListVideosParams) (responses [][]byte, err error) {
	for start := 0; err == nil; {
		var data []byte
		params.Start = start
		data, err = api.ListVideosRaw(params)
		if err != nil {
			if errors.Is(ErrorNoMoreResults, err) {
				return responses, nil
			}
			return responses, err
		}
		responses = append(responses, data)
		start = start + params.Count
	}
	return nil, errors.New("UNEXPECTED BEHAVIOUR in ListAllVideosRaw")
}

const apiVersion = "v1/"
const apiPrefix = "/api/" + apiVersion

// NewApiClient creates an authenticated API client by obtaining an access token.
//
// Parameters:
//   - clientID, clientSecret: OAuth credentials
//   - username, password: User authentication details
//   - Host: API endpoint hostname
//   - Protocol: Network protocol (e.g., "https")
//   - RateLimit: Optional Map from endpoint path (eg "/api/v1/videos") to a RateLimit struct that controls the limits.
//   - doRequest: Optional custom HTTP request handler
//
// Returns an initialized ApiClient and any error encountered during authentication.
func NewApiClient(clientID, clientSecret, username, password, Host, Protocol string, RateLimit RateLimitMap, doRequest *func(req *http.Request) (response *http.Response, err error)) (client *ApiClient, err error) {
	if doRequest == nil { // enable us to do web requests
		// this can be used to add proxies or do rate limiting.
		doRequestCopy := http.DefaultClient.Do
		doRequest = &doRequestCopy
	}

	header := http.Header{
		"Content-Type": []string{"application/x-www-form-urlencoded"},
		"User-Agent":   []string{"peertube-stats"},
	}
	request := *doRequest

	loginParams := url.URL{
		Scheme:     Protocol,
		Host:       Host,
		Path:       apiPrefix + "users/token",
		ForceQuery: true,
	}
	loginQuery := url.Values{}
	loginQuery.Set("client_id", clientID)
	loginQuery.Set("client_secret", clientSecret)
	loginQuery.Set("username", username)
	loginQuery.Set("password", password)
	loginQuery.Set("grant_type", "password")
	loginQuery.Set("response_type", "code")

	var expectedResponse = tokenLoginData{}

	loginResponse, err := request(&http.Request{
		Method: http.MethodPost,
		URL:    &loginParams,
		Header: header,
		Body:   io.NopCloser(strings.NewReader(loginQuery.Encode())),
	})
	if err != nil {
		return nil, errors.Join(errors.New("API login http request failed"), err)
	}

	defer loginResponse.Body.Close()
	if loginResponse.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(loginResponse.Body)
		return nil, errors.Join(errors.New("API login failed"), errors.New(string(responseBody)))
	}
	err = json.NewDecoder(loginResponse.Body).Decode(&expectedResponse)
	if err != nil {
		return nil, err
	}
	// Change the headers to use the obtained login
	header.Add("Authorization", expectedResponse.TokenType+" "+expectedResponse.AccessToken)

	if RateLimit != nil { // don't hook if there is no rate limit configured.
		// hook the request function to do Rate limiting
		requestCopy := request
		request = func(req *http.Request) (response *http.Response, err error) {
			var limit = RateLimit.Match(endpointPath(strings.TrimPrefix(req.URL.Path, Protocol+"://"+Host+apiPrefix)))
			if limit != nil { // if no rate limit is configured, this is skipped
				limit.Request()
			} else {
				LogHelp.NewLog(LogHelp.Warn, "Failed to find Rate limit rules", map[string]interface{}{"path": req.URL.Path, "rateLimit": RateLimit}).Log()
			}
			return requestCopy(req)

		}
	}

	client = &ApiClient{
		clientId:     clientID,
		clientSecret: clientSecret,
		Host:         Host,
		Protocol:     Protocol,
		doRequest:    request, // hooked with rate limiting
		headers:      header,
		accessToken:  expectedResponse.AccessToken,
		tokenData:    &expectedResponse,
		isAdmin:      username == "admin" || username == "root" || username == "administrator",
	}

	if RateLimit != nil {
		client.RateLimit = RateLimit
	}

	return client, nil
}
