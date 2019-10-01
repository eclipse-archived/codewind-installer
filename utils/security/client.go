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

// RegisteredClients : A collection of registered clients
type RegisteredClients struct {
	Collection []RegisteredClient
}

// RegisteredClient : Registered client
type RegisteredClient struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
	Name     string `json:"name"`
}

// RegisteredClientSecret : Client secret
type RegisteredClientSecret struct {
	Type   string `json:"type"`
	Secret string `json:"value"`
}

// SecClientCreate : Create a new client in Keycloak
func SecClientCreate(c *cli.Context) *SecError {
	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))
	clientid := strings.TrimSpace(c.String("clientid"))
	redirect := strings.TrimSpace(c.String("redirect"))

	// authenticate if needed
	if accesstoken == "" {
		authToken, err := SecAuthenticate(c, KeycloakMasterRealm, KeycloakAdminClientID)
		if err != nil || authToken == nil {
			return err
		}
		accesstoken = authToken.AccessToken
	}

	// build REST request
	url := hostname + "/auth/admin/realms/" + realm + "/clients"
	callbackRedirect := ""
	if redirect != "" {
		callbackRedirect = ",\"redirectUris\":[\"" + redirect + "\"]"
	}
	payload := strings.NewReader("{\"directAccessGrantsEnabled\":true, \"publicClient\":true, \"clientId\":\"" + clientid + "\",\"name\":\"" + clientid + "\"" + callbackRedirect + "}")
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
	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != "" {
		keycloakAPIError := parseKeycloakError(string(body), res.StatusCode)
		keycloakAPIError.Error = errOpResponseFormat
		kcError := errors.New(string(keycloakAPIError.ErrorDescription))
		return &SecError{keycloakAPIError.Error, kcError, kcError.Error()}
	}
	return nil
}

// SecClientGet : Retrieve Client information
func SecClientGet(c *cli.Context) (*RegisteredClient, *SecError) {
	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))
	clientid := strings.TrimSpace(c.String("clientid"))

	// authenticate if needed
	if accesstoken == "" {
		authToken, err := SecAuthenticate(c, KeycloakMasterRealm, KeycloakAdminClientID)
		if err != nil || authToken == nil {
			return nil, err
		}
		accesstoken = authToken.AccessToken
	}

	// build REST request
	url := hostname + "/auth/admin/realms/" + realm + "/clients?clientId=" + clientid
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

	// handle HTTP status codes
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		err = errors.New(string(body))
		return nil, &SecError{errOpResponse, err, err.Error()}
	}

	registeredClients := RegisteredClients{}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal([]byte(body), &registeredClients.Collection)
	if err != nil {
		return nil, &SecError{errOpResponseFormat, err, err.Error()}
	}

	registeredClient := RegisteredClient{}
	if len(registeredClients.Collection) > 0 {
		registeredClient = registeredClients.Collection[0]
		return &registeredClient, nil
	}

	return nil, nil
}

// SecClientGetSecret : Retrieve the client secret for the supplied clientID
func SecClientGetSecret(c *cli.Context) (*RegisteredClientSecret, *SecError) {

	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))

	// authenticate if needed
	if accesstoken == "" {
		authToken, err := SecAuthenticate(c, KeycloakMasterRealm, KeycloakAdminClientID)
		if err != nil || authToken == nil {
			return nil, err
		}
		accesstoken = authToken.AccessToken
	}

	registeredClient, secError := SecClientGet(c)
	if secError != nil {
		return nil, secError
	}

	if registeredClient == nil {
		return nil, nil
	}

	// build REST request
	url := hostname + "/auth/admin/realms/" + realm + "/clients/" + registeredClient.ID + "/client-secret"
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

	// handle HTTP status codes
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		err = errors.New(string(body))
		return nil, &SecError{errOpResponse, err, err.Error()}
	}

	registeredClientSecret := RegisteredClientSecret{}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal([]byte(body), &registeredClientSecret)
	if err != nil {
		return nil, &SecError{errOpResponseFormat, err, err.Error()}
	}

	return &registeredClientSecret, nil
}
