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

package sechttp

import (
	"errors"
	"flag"
	"net/http"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/zalando/go-keyring"
)

// DispatchHTTPRequest : Perform an HTTP request against PFE with token based authentication
// Returns: HTTPResponse, HTTPSecError
func DispatchHTTPRequest(httpClient utils.HTTPClient, originalRequest *http.Request, username string, connectionID string) (*http.Response, *HTTPSecError) {

	logr.Tracef("Request URL: %v %v\n", originalRequest.Method, originalRequest.URL)

	if strings.ToLower(connectionID) == "local" {
		response, err := sendRequest(httpClient, originalRequest, "")
		if err == nil {
			logr.Tracef("Received HTTP Status code: %v\n", response.StatusCode)
			return response, nil
		}
	}

	// Should be a 401 (bearer only) but is infact a 302 (Redirect to a login page)
	keycloakLoginErrorStatus := http.StatusFound
	logr.Tracef("Getting Connection: %v\n", connectionID)

	// Get the remote connection details
	con, conErr := connections.GetConnectionByID(connectionID)
	if conErr != nil {
		return nil, &HTTPSecError{errOpNoConnection, conErr.Err, conErr.Desc}
	}

	// Get the current access token from the keychain
	logr.Traceln("Retrieving an access token from the keychain")
	conID := strings.TrimSpace(strings.ToLower(connectionID))
	accessToken, _ := keyring.Get(security.KeyringServiceName+"."+conID, "access_token")

	if accessToken == "" {
		logr.Traceln("Access token not found in keychain")
	} else {
		logr.Traceln("Access token found in keychain, trying request")
		response, err := sendRequest(httpClient, originalRequest, accessToken)
		if err == nil && response.StatusCode != keycloakLoginErrorStatus {
			logr.Tracef("Received HTTP Status code: %v", response.StatusCode)
			return response, nil
		}
		logr.Tracef(" Request failed: %v", err.Desc)
	}

	// Try refreshing the access token with our cached refresh token
	logr.Tracef("Retrieving a refresh token from the keychain")
	refreshToken, _ := keyring.Get(security.KeyringServiceName+"."+conID, "refresh_token")
	if refreshToken == "" {
		logr.Tracef("Refresh token not found in keychain")
	} else {
		logr.Tracef("Try refreshing the access token with our cached refresh token")
		tokens, secError := security.SecRefreshAccessToken(http.DefaultClient, con, refreshToken)
		if secError != nil {
			logr.Tracef("Failed refreshing access token %v : %v\n", secError.Op, secError.Desc)
		}
		if tokens != nil {
			logr.Tracef("New access token received")
			accessToken = tokens.AccessToken
			logr.Tracef("Trying the original request again with the new access_token")
			response, err := sendRequest(httpClient, originalRequest, accessToken)
			if err == nil && response.StatusCode != keycloakLoginErrorStatus {
				logr.Tracef("Received HTTP Status code: %v", response.StatusCode)
				return response, nil
			}
		}
	}

	logr.Tracef("Re-authenticate using cached credentials from the keychain")
	password, keyErr := keyring.Get(security.KeyringServiceName+"."+conID, strings.ToLower(username))
	if keyErr != nil {
		logr.Tracef("ERROR:  %v\n", keyErr.Error())
		err := errors.New(errMissingPassword)
		return nil, &HTTPSecError{errOpNoPassword, err, err.Error()}
	}

	set := flag.NewFlagSet("Authentication", 0)
	set.String("host", con.AuthURL, "doc")
	set.String("realm", con.Realm, "doc")
	set.String("username", username, "doc")
	set.String("password", password, "doc")
	set.String("client", con.ClientID, "doc")
	set.String("conid", con.ID, "doc")
	c := cli.NewContext(nil, set, nil)
	tokens, secError := security.SecAuthenticate(http.DefaultClient, c, "", "")
	if secError != nil {
		// Bailing out, user cant authenticate
		logr.Tracef("Bailing out, user can not authenticate")
		return nil, &HTTPSecError{errOpAuthFailed, secError.Err, secError.Desc}
	}

	// Try to access the resource again with the new access token
	logr.Tracef("Try to access the resource again with the new access token")
	response, err := sendRequest(httpClient, originalRequest, tokens.AccessToken)

	if err == nil {
		logr.Tracef("Received HTTP Status code: %v", response.StatusCode)
		return response, nil
	}

	// No other methods of authentication left to try, tell the user and give up
	logr.Tracef("No other methods of authentication left to try, tell the user and give up")
	failedError := errors.New("No other methods left to try")
	return nil, &HTTPSecError{errOpFailed, failedError, failedError.Error()}
}

// Send the HTTP request along with supplied headers and access_token
func sendRequest(httpClient utils.HTTPClient, originalRequest *http.Request, accessToken string) (*http.Response, *HTTPSecError) {

	// Add auth headers
	if accessToken != "" {
		originalRequest.Header.Set("Authorization", "bearer "+accessToken)
		originalRequest.Header.Set("Cache-Control", "no-cache")
		originalRequest.Header.Set("cache-control", "no-cache")
	}

	// send request
	res, err := httpClient.Do(originalRequest)
	if err != nil {
		logr.Tracef("sendRequest: REQUEST FAILED")
		return nil, &HTTPSecError{errOpNoConnection, err, err.Error()}
	}
	return res, nil
}
