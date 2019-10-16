/*******************************************************************************
 * Copyright (c) 2019 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

package security

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/eclipse/codewind-installer/utils"
	"github.com/eclipse/codewind-installer/utils/deployments"
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
func SecAuthenticate(httpClient utils.HTTPClient, c *cli.Context, connectionRealm string, connectionClient string) (*AuthToken, *SecError) {

	cliHostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	cliUsername := strings.TrimSpace(strings.ToLower(c.String("username")))
	cliRealm := strings.TrimSpace(strings.ToLower(c.String("realm")))
	cliClient := strings.TrimSpace(strings.ToLower(c.String("client")))
	cliPassword := strings.TrimSpace(c.String("password"))
	deploymentID := strings.TrimSpace(strings.ToLower(c.String("depid")))

	// Check supplied context flags
	if deploymentID == "" && (cliHostname == "" || cliUsername == "" || cliRealm == "" || cliClient == "") {
		err := errors.New("Must supply a deployment ID or connection details")
		return nil, &SecError{errOpDepConfig, err, err.Error()}
	}

	hostname := ""
	username := ""
	password := ""
	realm := ""
	client := ""

	// Check deployment is known
	deployment, depErr := deployments.GetDeploymentByID(deploymentID)
	if deploymentID != "" && depErr != nil {
		return nil, &SecError{errOpDepConfig, depErr.Err, depErr.Desc}
	}

	if deployment != nil {
		hostname = deployment.AuthURL
		realm = deployment.Realm
		client = deployment.ClientID
	}

	// Use command line context flags in preference to loaded deployment fields
	if cliHostname != "" {
		hostname = cliHostname
	}
	if cliUsername != "" {
		username = cliUsername
	}
	if cliRealm != "" {
		realm = cliRealm
	}
	if cliClient != "" {
		client = cliClient
	}

	// When a matching deployment exist retrieve secret from the keyring
	if deployment != nil {
		secret, secError := SecKeyGetSecret(deployment.ID, username)
		if secError != nil && cliPassword == "" {
			return nil, secError
		}
		password = secret
	}

	if cliPassword != "" {
		password = cliPassword
	}

	// If a connection realm was supplied, use that instead of the command line Context flags.
	// This allows this function to be used by other realms such as master when admins are performing initial setup of keycloak
	if connectionRealm != "" {
		realm = connectionRealm
	}

	if connectionClient != "" {
		client = connectionClient
	}

	// Pre-flight check

	if hostname == "" || realm == "" || username == "" || password == "" || client == "" {
		err := errors.New(textInvalidOptions)
		return nil, &SecError{errOpCLICommand, err, err.Error()}
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
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, &SecError{errOpConnection, err, err.Error()}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	// Handle special case http status codes
	switch httpCode := res.StatusCode; {
	case httpCode == http.StatusBadRequest, httpCode == http.StatusUnauthorized:
		keycloakAPIError := parseKeycloakError(string(body), res.StatusCode)
		kcError := errors.New(string(keycloakAPIError.ErrorDescription))
		return nil, &SecError{keycloakAPIError.Error, kcError, kcError.Error()}
	case httpCode != http.StatusOK:
		err = errors.New(string(body))
		return nil, &SecError{errOpResponse, err, err.Error()}
	}

	// Parse and return authtoken
	authToken := AuthToken{}
	err = json.Unmarshal([]byte(body), &authToken)
	if err != nil {
		return nil, &SecError{errOpResponseFormat, err, textUnableToParse}
	}

	// store access and refresh tokens in keyring if a deployment is known
	if deployment != nil {
		secErr := SecKeyUpdate(deploymentID, "access_token", authToken.AccessToken)
		if secErr != nil {
			return &authToken, secErr
		}
		secErr = SecKeyUpdate(deploymentID, "refresh_token", authToken.RefreshToken)
		if secErr != nil {
			return &authToken, secErr
		}

		// login successful, update users password in keyring
		if password != "" {
			secErr = SecKeyUpdate(deploymentID, username, password)
			if secErr != nil {
				return &authToken, secErr
			}
		}
	}

	return &authToken, nil
}
