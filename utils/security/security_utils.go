package security

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/urfave/cli"
)

// KeycloakMasterRealm master realm name
const KeycloakMasterRealm string = "master"

// KeycloakAdminClientID master realm name
const KeycloakAdminClientID string = "admin-cli"

// CodewindCliID master realm name
const CodewindCliID string = "codewind-cli"

// CodewindClientID master realm name
const CodewindClientID string = "codewind-backend"

// SecError - Security package errors
type SecError struct {
	Op   string
	Err  error
	Desc string
}

const (
	errOpConnection     = "sec_connection" // Connection failed
	errOpResponse       = "sec_response"   // Bad response
	errOpResponseFormat = "sec_bodyparser" // Parse errors
	errOpNotFound       = "sec_notfound"   // No matching search results
)

// SecError : Error formatted in JSON containing an errorOp and a description from
// either a fault condition in the CLI, or an error payload from a REST request
func (se *SecError) Error() string {
	return "{\"error\", \"" + se.Op + "\", \"error_description\": \"" + se.Err.Error() + "\"}"
}

// RegisteredClients A collection of registered clients
type RegisteredClients struct {
	Collection []RegisteredClient
}

// RegisteredClient details of a registered client
type RegisteredClient struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
	Name     string `json:"name"`
}

// KeycloakAPIError Error responses from Keycloak
type KeycloakAPIError struct {
	HTTPStatus       int
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorMessage     string `json:"errorMessage"`
}

// Result : status messaqe
type Result struct {
	Status string `json:"status"`
}

// parseKeycloakError : parse the JSON response from Keycloak
func parseKeycloakError(body string, httpCode int) *KeycloakAPIError {
	keycloakAPIError := KeycloakAPIError{}
	keycloakAPIError.HTTPStatus = httpCode
	json.Unmarshal([]byte(body), &keycloakAPIError)
	// copy message into description if one is not set
	if keycloakAPIError.ErrorMessage != "" && keycloakAPIError.ErrorDescription == "" {
		keycloakAPIError.ErrorDescription = keycloakAPIError.ErrorMessage
	}
	return &keycloakAPIError
}

// SecUserCreate : Create a new user in Keycloak
func SecUserCreate(c *cli.Context) error {
	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return nil
}

// SecUserGet : Retrieve user info from keycloak
func SecUserGet(c *cli.Context) error {
	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return nil
}

// SecUserSetPW : Set a users password
func SecUserSetPW(c *cli.Context) error {
	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return nil
}
