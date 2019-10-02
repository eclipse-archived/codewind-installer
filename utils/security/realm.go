package security

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/urfave/cli"
)

// SecRealmCreate : Create a new realm in Keycloak
func SecRealmCreate(c *cli.Context) *SecError {

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))

	// Authenticate if needed
	if accesstoken == "" {
		authToken, err := SecAuthenticate(c, KeycloakMasterRealm, KeycloakAdminClientID)
		if err != nil || authToken == nil {
			return err
		}
		accesstoken = authToken.AccessToken
	}

	// build REST request
	url := hostname + "/auth/admin/realms"

	// build the payload (JSON)
	type PayloadRealm struct {
		Realm               string `json:"realm"`
		DisplayName         string `json:"displayName"`
		Enabled             bool   `json:"enabled"`
		LoginTheme          string `json:"loginTheme"`
		AccessTokenLifespan int    `json:"accessTokenLifespan"`
	}
	tempRealm := &PayloadRealm{
		Realm:               realm,
		DisplayName:         realm,
		Enabled:             true,
		LoginTheme:          "codewind",
		AccessTokenLifespan: 86400,
	}

	jsonRealm, err := json.Marshal(tempRealm)
	payload := strings.NewReader(string(jsonRealm))
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return &SecError{errOpConnection, err, err.Error()}
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Authorization", "Bearer "+accesstoken)

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &SecError{errOpConnection, err, err.Error()}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if string(body) != "" {
		keycloakAPIError := parseKeycloakError(string(body), res.StatusCode)
		keycloakAPIError.Error = errOpResponseFormat
		kcError := errors.New(keycloakAPIError.ErrorDescription)
		return &SecError{keycloakAPIError.Error, kcError, kcError.Error()}
	}
	return nil
}
