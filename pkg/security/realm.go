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

	"github.com/urfave/cli"
)

// KeycloakRealm : A Keycloak Realm
type KeycloakRealm struct {
	ID          string `json:"id"`
	Realm       string `json:"realm"`
	DisplayName string `json:"displayName"`
	Enabled     bool   `json:"enabled"`
	LoginTheme  string `json:"loginTheme"`
}

// SecRealmCreate : Create a new realm in Keycloak
func SecRealmCreate(c *cli.Context) *SecError {

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	newRealm := strings.TrimSpace(c.String("newrealm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))

	themeToUse, secErr := GetSuggestedTheme(hostname, accesstoken)
	if secErr != nil {
		return secErr
	}

	// build REST request
	url := hostname + "/auth/admin/realms"

	// build the payload (JSON)
	type PayloadRealm struct {
		Realm                 string `json:"realm"`
		DisplayName           string `json:"displayName"`
		Enabled               bool   `json:"enabled"`
		LoginTheme            string `json:"loginTheme"`
		AccessTokenLifespan   int    `json:"accessTokenLifespan"`
		SSOSessionIdleTimeout int    `json:"ssoSessionIdleTimeout"`
		SSOSessionMaxLifespan int    `json:"ssoSessionMaxLifespan"`
	}
	tempRealm := &PayloadRealm{
		Realm:                 newRealm,
		DisplayName:           newRealm,
		Enabled:               true,
		LoginTheme:            themeToUse,
		AccessTokenLifespan:   (1 * 24 * 60 * 60), // access tokens last 1 day
		SSOSessionIdleTimeout: (5 * 24 * 60 * 60), // refresh tokens last 5 days
		SSOSessionMaxLifespan: (5 * 24 * 60 * 60), // refresh tokens last 5 days
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

// SecRealmGet : Reads a realm in Keycloak
func SecRealmGet(authURL string, accessToken string, realmName string) (*KeycloakRealm, *SecError) {

	req, err := http.NewRequest("GET", authURL+"/auth/admin/realms/"+realmName, nil)
	if err != nil {
		return nil, &SecError{errOpConnection, err, err.Error()}
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Authorization", "Bearer "+accessToken)
	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &SecError{errOpConnection, err, err.Error()}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if res.StatusCode == http.StatusOK {
		// Parse Realm
		keycloakRealm := KeycloakRealm{}
		err = json.Unmarshal([]byte(body), &keycloakRealm)
		if err != nil {
			kcError := errors.New("Error parsing")
			return nil, &SecError{errOpResponseFormat, kcError, kcError.Error()}
		}
		return &keycloakRealm, nil
	}

	if string(body) != "" {
		keycloakAPIError := parseKeycloakError(string(body), res.StatusCode)
		keycloakAPIError.Error = errOpResponseFormat
		kcError := errors.New(keycloakAPIError.ErrorDescription)
		return nil, &SecError{keycloakAPIError.Error, kcError, kcError.Error()}
	}

	return nil, nil
}
