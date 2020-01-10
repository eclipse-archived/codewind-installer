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

	cwerrors "github.com/eclipse/codewind-installer/pkg/errors"
)

// RegisteredTheme : A Keycloak theme
type RegisteredTheme struct {
	Name    string   `json:"name"`
	Locales []string `json:"locales"`
}

// RegisteredThemes : A collection of themes
type RegisteredThemes struct {
	Common  []RegisteredTheme `json:"common"`
	Admin   []RegisteredTheme `json:"admin"`
	Login   []RegisteredTheme `json:"login"`
	Welcome []RegisteredTheme `json:"welcome"`
	Account []RegisteredTheme `json:"account"`
	Email   []RegisteredTheme `json:"email"`
}

// ServerInfo : A collection of themes
type ServerInfo struct {
	Themes RegisteredThemes `json:"themes"`
}

// GetServerInfo - fetch Keycloak server info
func GetServerInfo(keycloakHostname string, accesstoken string) (*ServerInfo, *cwerrors.BasicError) {

	// build REST request
	url := keycloakHostname + "/auth/admin/serverinfo"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+accesstoken)
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("cache-control", "no-cache")

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, &cwerrors.BasicError{errOpResponseFormat, err, err.Error()}
	}

	// Handle special case http status codes
	switch httpCode := res.StatusCode; {
	case httpCode == http.StatusBadRequest, httpCode == http.StatusUnauthorized:
		keycloakAPIError := parseKeycloakError(string(body), res.StatusCode)
		kcError := errors.New(string(keycloakAPIError.ErrorDescription))
		return nil, &cwerrors.BasicError{keycloakAPIError.Error, kcError, kcError.Error()}
	case httpCode != http.StatusOK:
		err = errors.New(string(body))
		return nil, &cwerrors.BasicError{errOpResponse, err, err.Error()}
	}

	// Parse and return ServerInfo
	serverInfo := ServerInfo{}
	err = json.Unmarshal([]byte(body), &serverInfo)
	if err != nil {
		return nil, &cwerrors.BasicError{errOpResponseFormat, err, textUnableToParse}
	}
	return &serverInfo, nil
}

// GetSuggestedTheme - Recommends the Codewind theme, else Che, else keycloak default
func GetSuggestedTheme(keycloakHostname string, accesstoken string) (string, *cwerrors.BasicError) {
	serverInfo, secErr := GetServerInfo(keycloakHostname, accesstoken)
	if secErr != nil {
		return "", secErr
	}

	loginThemes := serverInfo.Themes.Login
	if len(loginThemes) == 0 {
		return "", nil
	}

	themeCodewind := ""
	themeChe := ""
	themeKeycloak := ""

	for _, theme := range loginThemes {
		switch strings.ToLower(theme.Name) {
		case "codewind":
			{
				themeCodewind = theme.Name
				break
			}
		case "che":
			{
				themeChe = theme.Name
				break
			}
		case "keycloak":
			{
				themeKeycloak = theme.Name
				break
			}
		}
	}

	if themeCodewind != "" {
		return themeCodewind, nil
	}
	if themeChe != "" {
		return themeChe, nil
	}
	if themeKeycloak != "" {
		return themeKeycloak, nil
	}
	return "", nil

}
