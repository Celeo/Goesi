package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"net/url"
)

var log = logging.MustGetLogger("goesi")

// ESI is the interface for interacting with the EVE Swagger Interface
type ESI struct {
	client            *http.Client
	Version           string
	ClientID          string
	ClientSecret      string
	ClientCallbackURL string
	UserAgent         string
	Scope             string
	AccessToken       string
	RefreshToken      string
}

var (
	// BaseURL is the top-level URL of ESI
	BaseURL = "https://esi.tech.ccp.is/"
	// OauthURL is the URL for making the first OAuth request
	OauthURL = "https://login.eveonline.com/oauth/"
	// TokenURL is the URL for making the call to exchange Oauth code for a token
	TokenURL = "https://login.eveonline.com/oauth/token"
	// VerifyURL is the "whoami" URL
	VerifyURL = "https://login.eveonline.com/oauth/verify"
	// AuthorizeURL is the URL to generate the URL to send to the user
	AuthorizeURL = "https://login.eveonline.com/oauth/authorize"
)

// New creates a new instance of the ESI struct and returns it
func New(clientID, clientSecret, clientCallbackURL string) ESI {
	log.Debug("Initializing a new ESI struct")
	return ESI{
		&http.Client{},
		"latest",
		clientID,
		clientSecret,
		clientCallbackURL,
		"github.com/Celeo/Goesi",
		"",
		"",
		"",
	}
}

// GetAuthorizeURL returns the URL that a user must visit in order to authenticate with the SSO
func (e *ESI) GetAuthorizeURL() (string, error) {
	log.Debug("Creating authorization url")
	if e.ClientID == "" || e.ClientSecret == "" || e.ClientCallbackURL == "" {
		es := "Missing client data - cannot generate callback URL"
		log.Error(es)
		return "", fmt.Errorf(es)
	}
	return fmt.Sprintf("%s?response_type=code&redirect_uri=%s&client_id=%s&scope=%s",
		AuthorizeURL,
		e.ClientCallbackURL,
		e.ClientID,
		e.Scope,
	), nil
}

type authenticateResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// createAuthorizationHeader returns the header string required for getting an access token from SSO
func createAuthorizationHeader(e *ESI) string {
	header := "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", e.ClientID, e.ClientSecret)))
	log.Debug("Header: " + header)
	return header
}

// Authenticate takes a code from the SSO and fetches the access token
func (e *ESI) Authenticate(code string) error {
	log.Debug("Starting authorization flow")
	form := url.Values{
		"grant_type": []string{"authorization_code"},
		"code":       []string{code},
	}
	req, err := http.NewRequest("POST", TokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		log.Error("Cannot create a new request stuct")
		return err
	}
	req.Header.Add("Authorization", createAuthorizationHeader(e))
	req.Header.Add("User-Agent", e.UserAgent)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	log.Debug("Making request to authorization url")
	resp, err := e.client.Do(req)
	if err != nil {
		log.Error("Error making authorization url request")
		return err
	}

	log.Debug("Getting response")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Cannot read response body")
		return err
	}
	defer resp.Body.Close()
	if string(body) == "" || resp.StatusCode != http.StatusOK {
		log.Errorf("Error with authenticate response, code %d, body: '%s'", resp.StatusCode, body)
		log.Debug("Headers:")
		for key, value := range resp.Header {
			log.Debugf("%s = %s", key, value)
		}
		return fmt.Errorf("Response body is empty")
	}
	log.Debug("Parsing response")
	var respData authenticateResponse
	err = json.Unmarshal(body, &respData)
	if err != nil {
		log.Errorf("Error parsing response, body: '%s'", body)
		return err
	}

	log.Debug("Setting token")
	e.AccessToken = respData.AccessToken
	e.RefreshToken = respData.RefreshToken
	return nil
}

// WhoAmI returns basic information about the access token's character
func (e *ESI) WhoAmI() (*gabs.Container, error) {
	log.Info("Making whoami request")
	req, err := http.NewRequest("GET", VerifyURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", e.UserAgent)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+e.AccessToken)
	resp, err := e.client.Do(req)
	if err != nil {
		log.Error("Error making whoami request to ESI")
		return nil, err
	}
	defer resp.Body.Close()
	json, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		log.Error("Error converting response body to Gabs container")
		return nil, err
	}
	return json, nil
}

// Get fetches data from ESI
func (e *ESI) Get(path string) (*gabs.Container, error) {
	url := BaseURL + e.Version + "/" + path + "/"
	log.Info("Making call to URL '%s'\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Error creating a new request struct")
		return nil, err
	}
	if e.AccessToken != "" {
		req.Header.Add("Authorization", "Bearer"+e.AccessToken)
	}
	req.Header.Add("User-Agent", e.UserAgent)
	req.Header.Add("Accept", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		log.Error("Error making request to ESI")
		return nil, err
	}
	defer resp.Body.Close()
	json, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		log.Error("Error converting response body to Gabs container")
		return nil, err
	}
	return json, nil
}
