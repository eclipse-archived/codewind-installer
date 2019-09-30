package security

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/urfave/cli"
)

// RegisteredUsers A collection of registered users
type RegisteredUsers struct {
	Collection []RegisteredUser
}

// RegisteredUser details of a registered user
type RegisteredUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// SecUserSetPW : Set a users password
func SecUserSetPW(c *cli.Context) error {
	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return nil
}

// SecUserCreate : Create a new realm in Keycloak
func SecUserCreate(c *cli.Context) *SecError {
	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))
	targetUsername := strings.TrimSpace(c.String("name"))

	// Authenticate if needed
	if accesstoken == "" {
		authToken, err := SecAuthenticate(c, KeycloakMasterRealm, KeycloakAdminClientID)
		if err != nil || authToken == nil {
			return err
		}
		accesstoken = authToken.AccessToken
	}

	// build REST request
	url := hostname + "/auth/admin/realms/" + realm + "/users"
	//    '{"username":"developer","firstName":"codewind","lastName":"developer","enabled":true}'
	payload := strings.NewReader("{\"enabled\":true,\"username\":\"" + targetUsername + "\",\"firstName\":\"\",\"lastName\":\"" + targetUsername + "\"}")
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
		keycloakAPIError.Error = errOpCreate
		kcError := errors.New(keycloakAPIError.ErrorDescription)
		return &SecError{keycloakAPIError.Error, kcError, kcError.Error()}
	}
	return nil
}

// SecUserGet : Get user from Keycloak
func SecUserGet(c *cli.Context) (*RegisteredUser, *SecError) {
	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))
	searchName := strings.TrimSpace(c.String("name"))

	// Authenticate if needed
	if accesstoken == "" {
		authToken, err := SecAuthenticate(c, KeycloakMasterRealm, KeycloakAdminClientID)
		if err != nil || authToken == nil {
			return nil, err
		}
		accesstoken = authToken.AccessToken
	}

	// Built REST request
	url := hostname + "/auth/admin/realms/" + realm + "/users?username=" + searchName
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &SecError{errOpConnection, err, err.Error()}
	}
	req.Header.Add("Authorization", "Bearer "+accesstoken)
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &SecError{errOpConnection, err, err.Error()}
	}

	defer res.Body.Close()

	// Handle HTTP status codes
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		err = errors.New(string(body))
		return nil, &SecError{errOpResponse, err, err.Error()}
	}

	registeredUsers := RegisteredUsers{}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal([]byte(body), &registeredUsers.Collection)
	if err != nil {
		return nil, &SecError{errOpResponseFormat, err, err.Error()}
	}

	registredUser := RegisteredUser{}

	if len(registeredUsers.Collection) > 0 {
		registredUser = registeredUsers.Collection[0]
		return &registredUser, nil
	}

	return nil, nil
}
