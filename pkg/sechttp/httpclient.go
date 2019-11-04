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

// DispatchHTTPRequest : perform an HTTP request with token based authentication
// Returns: HTTPResponse, HTTPSecError
func DispatchHTTPRequest(httpClient utils.HTTPClient, originalRequest *http.Request, username string, connectionID string) (*http.Response, *HTTPSecError) {

	logr.SetLevel(logr.TraceLevel)

	logr.Printf("Request URL: %v %v", originalRequest.Method, originalRequest.URL)

	if strings.ToLower(connectionID) == "local" {
		response, err := sendRequest(httpClient, originalRequest, "")
		if err == nil {
			logr.Debugf("Received HTTP Status code: %v", response.StatusCode)
			return response, nil
		}
	}

	// Should be a 401 (bearer only) but is infact a 302 (Redirect to a login page)
	keycloakLoginErrorStatus := http.StatusFound
	logr.Debugf("Getting Connection: %v\n", connectionID)

	// Get the remote connection details
	con, conErr := connections.GetConnectionByID(connectionID)
	if conErr != nil {
		return nil, &HTTPSecError{errOpNoConnection, conErr.Err, conErr.Desc}
	}

	// Get the current access token from the keychain
	logr.Debugf("Retrieving an access token from the keychain")
	conID := strings.TrimSpace(strings.ToLower(connectionID))
	accessToken, _ := keyring.Get(security.KeyringServiceName+"."+conID, "access_token")

	if accessToken == "" {
		logr.Debugf("Access token not found in keychain")
	} else {
		logr.Debugf("Access token found in keychain, trying request")
		response, err := sendRequest(httpClient, originalRequest, accessToken)
		if err == nil && response.StatusCode != keycloakLoginErrorStatus {
			logr.Debugf("Received HTTP Status code: %v", response.StatusCode)
			return response, nil
		}
		logr.Debugf(" Request failed: %v", err.Desc)
	}

	// Try refreshing the access token with our cached refresh token
	logr.Debugf("Retrieving a refresh token from the keychain")
	refreshToken, _ := keyring.Get(security.KeyringServiceName+"."+conID, "refresh_token")
	if refreshToken == "" {
		logr.Debugf("Refresh token not found in keychain")
	} else {
		logr.Debugf("Try refreshing the access token with our cached refresh token")
		tokens, secError := security.SecRefreshAccessToken(http.DefaultClient, con, refreshToken)
		if secError != nil {
			logr.Debugf("Failed refreshing access token %v : %v\n", secError.Op, secError.Desc)
		}
		if tokens != nil {
			logr.Debugf("New access token received")
			accessToken = tokens.AccessToken
			logr.Debugf("Trying the original request again with the new access_token")
			response, err := sendRequest(httpClient, originalRequest, accessToken)
			if err == nil && response.StatusCode != keycloakLoginErrorStatus {
				logr.Debugf("Received HTTP Status code: %v", response.StatusCode)
				return response, nil
			}
		}
	}

	logr.Debugf("Re-authenticate using cached credentials from the keychain")
	password, keyErr := keyring.Get(security.KeyringServiceName+"."+conID, strings.ToLower(username))
	if keyErr != nil {
		logr.Debugf("ERROR:  %v\n", keyErr.Error())
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
		logr.Debugf("Bailing out, user can not authenticate")
		return nil, &HTTPSecError{errOpAuthFailed, secError.Err, secError.Desc}
	}

	// Try to access the resource again with the new access token
	logr.Debugf("Try to access the resource again with the new access token")
	response, err := sendRequest(httpClient, originalRequest, tokens.AccessToken)

	if err == nil {
		logr.Debugf("Received HTTP Status code: %v", response.StatusCode)
		return response, nil
	}

	// No other methods of authentication left to try, tell the user and give up
	logr.Debugf("No other methods of authentication left to try, tell the user and give up")
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
		logr.Debugf("sendRequest: REQUEST FAILED")
		return nil, &HTTPSecError{errOpNoConnection, err, err.Error()}
	}
	return res, nil
}
