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

// Role : Access role
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Composite   bool   `json:"composite"`
	ClientRole  bool   `json:"clientRole"`
	ContainerID string `json:"containerId"`
}

// SecRoleCreate : Create a new role in Keycloak
func SecRoleCreate(c *cli.Context) *SecError {

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realmName := strings.TrimSpace(c.String("realm"))
	roleName := strings.TrimSpace(c.String("role"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))

	// build REST request
	url := hostname + "/auth/admin/realms/" + realmName + "/roles"

	// Role : Access role
	type NewRole struct {
		Name        string `json:"name"`
		Composite   bool   `json:"composite"`
		ClientRole  bool   `json:"clientRole"`
		ContainerID string `json:"containerId"`
	}

	tempRole := &NewRole{
		Name:        roleName,
		Composite:   false,
		ClientRole:  false,
		ContainerID: realmName,
	}
	jsonRole, err := json.Marshal(tempRole)

	payload := strings.NewReader(string(jsonRole))
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

	if res.StatusCode != http.StatusCreated {
		secErr := errors.New("HTTP " + res.Status)
		return &SecError{errOpConnection, secErr, secErr.Error()}
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

func getRoleByName(c *cli.Context, roleName string) (*Role, *SecError) {

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))
	realmName := strings.TrimSpace(c.String("realm"))

	requestedRole := roleName
	if requestedRole == "" {
		requestedRole = strings.TrimSpace(c.String("role"))
	}
	// build REST request
	url := hostname + "/auth/admin/realms/" + realmName + "/roles/" + requestedRole
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &SecError{errOpConnection, err, err.Error()}
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Authorization", "Bearer "+accesstoken)

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &SecError{errOpConnection, err, err.Error()}
	}

	// check we received a valid response
	if res.StatusCode != http.StatusOK {
		unableToReadErr := errors.New("Bad response")
		return nil, &SecError{errOpConnection, unableToReadErr, unableToReadErr.Error()}
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	// parse the result
	var role *Role
	err = json.Unmarshal([]byte(body), &role)
	if err != nil {
		return nil, &SecError{errOpResponseFormat, err, textUnableToParse}
	}

	// found role
	return role, nil
}
