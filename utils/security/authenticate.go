package security

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/urfave/cli"
)

// AuthToken from the keycloak server after successfully authenticating
type AuthToken struct {
	AccessToken     string `json:"access_token"`
	ExpiresIn       int    `json:"expires_in"`
	RefreshToken    string `json:"refresh_token"`
	TokenType       string `json:"token_type"`
	NotBeforePolicy int    `json:"not-before-policy"`
	SessionState    string `json:"session_state"`
	Scope           string `json:"scope"`
}

// SecAuthenticate - sends credentials to the auth server for a specific realm and returns an AuthToken
// connectionRealm can be used to override the supplied context arguments
func SecAuthenticate(c *cli.Context, connectionRealm string, connectionClient string) (*AuthToken, *SecError) {

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	username := strings.TrimSpace(strings.ToLower(c.String("username")))
	password := strings.TrimSpace(c.String("password"))
	realm := strings.TrimSpace(strings.ToLower(c.String("realm")))
	client := strings.TrimSpace(strings.ToLower(c.String("client")))

	// If a connection realm was supplied, use that instead of the command line Context flags
	if connectionRealm != "" {
		realm = connectionRealm
	}

	if connectionClient != "" {
		client = connectionClient
	}

	// build REST request
	url := hostname + "/auth/realms/" + realm + "/protocol/openid-connect/token"
	payload := strings.NewReader("grant_type=password&client_id=" + client + "&username=" + username + "&password=" + password)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, &SecError{errOpConnection, err, err.Error()}
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("cache-control", "no-cache")

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &SecError{errOpConnection, err, err.Error()}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	// Handle special case http status codes
	switch httpCode := res.StatusCode; {
	case httpCode == 400, httpCode == 401:
		keycloakAPIError := parseKeycloakError(string(body), res.StatusCode)
		kcError := errors.New(string(keycloakAPIError.ErrorDescription))
		return nil, &SecError{keycloakAPIError.Error, kcError, kcError.Error()}
	case httpCode != 200:
		err = errors.New(string(body))
		return nil, &SecError{errOpResponse, err, err.Error()}
	}

	// Parse and return authtoken
	authToken := AuthToken{}
	err = json.Unmarshal([]byte(body), &authToken)
	if err != nil {
		return nil, &SecError{errOpResponseFormat, err, textUnableToParse}
	}
	return &authToken, nil
}
