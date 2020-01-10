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
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// RegisteredUsers : A collection of registered users
type RegisteredUsers struct {
	Collection []RegisteredUser
}

// RegisteredUser : details of a registered user
type RegisteredUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// SecUserCreate : Create a new realm in Keycloak
func SecUserCreate(c *cli.Context) *cwerrors.BasicError {

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))
	targetUsername := strings.TrimSpace(c.String("name"))

	// authenticate if needed
	if accesstoken == "" {
		authToken, err := SecAuthenticate(http.DefaultClient, c, KeycloakMasterRealm, KeycloakAdminClientID)
		if err != nil || authToken == nil {
			return err
		}
		accesstoken = authToken.AccessToken
	}

	// build REST request
	url := hostname + "/auth/admin/realms/" + realm + "/users"

	// build the payload (JSON)
	type PayloadUser struct {
		Enabled   bool   `json:"enabled"`
		Username  string `json:"username"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	tempUser := &PayloadUser{
		Enabled:   true,
		Username:  targetUsername,
		FirstName: "",
		LastName:  "",
	}

	jsonUser, err := json.Marshal(tempUser)
	payload := strings.NewReader(string(jsonUser))
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Authorization", "Bearer "+accesstoken)

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if string(body) != "" {
		keycloakAPIError := parseKeycloakError(string(body), res.StatusCode)
		keycloakAPIError.Error = errOpCreate
		kcError := errors.New(keycloakAPIError.ErrorDescription)
		return &cwerrors.BasicError{keycloakAPIError.Error, kcError, kcError.Error()}
	}
	return nil
}

// SecUserGet : Get user from Keycloak
func SecUserGet(c *cli.Context) (*RegisteredUser, *cwerrors.BasicError) {

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))
	searchName := strings.TrimSpace(c.String("name"))

	// authenticate if needed
	if accesstoken == "" {
		authToken, err := SecAuthenticate(http.DefaultClient, c, KeycloakMasterRealm, KeycloakAdminClientID)
		if err != nil || authToken == nil {
			return nil, err
		}
		accesstoken = authToken.AccessToken
	}

	// build REST request
	url := hostname + "/auth/admin/realms/" + realm + "/users?username=" + searchName
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}
	req.Header.Add("Authorization", "Bearer "+accesstoken)
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}

	defer res.Body.Close()

	// handle HTTP status codes
	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		err = errors.New(string(body))
		return nil, &cwerrors.BasicError{errOpResponse, err, err.Error()}
	}

	registeredUsers := RegisteredUsers{}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal([]byte(body), &registeredUsers.Collection)
	if err != nil {
		return nil, &cwerrors.BasicError{errOpResponseFormat, err, err.Error()}
	}

	registeredUser := RegisteredUser{}

	if len(registeredUsers.Collection) > 0 {
		registeredUser = registeredUsers.Collection[0]
		return &registeredUser, nil
	}

	// user not found
	errNotFound := errors.New(textUserNotFound)
	return nil, &cwerrors.BasicError{errOpNotFound, errNotFound, errNotFound.Error()}

}

// SecUserSetPW : Resets the users password in keycloak to a new one supplied
func SecUserSetPW(c *cli.Context) *cwerrors.BasicError {

	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))
	newPassword := strings.TrimSpace(c.String("newpw"))

	// authenticate if needed
	if accesstoken == "" {
		authToken, err := SecAuthenticate(http.DefaultClient, c, KeycloakMasterRealm, KeycloakAdminClientID)
		if err != nil || authToken == nil {
			return err
		}
		accesstoken = authToken.AccessToken
	}

	registeredUser, secError := SecUserGet(c)
	if secError != nil {
		return secError
	}

	// build REST request
	url := hostname + "/auth/admin/realms/" + realm + "/users/" + registeredUser.ID + "/reset-password"

	// build the payload (JSON)
	type PayloadPasswordChange struct {
		Type      string `json:"type"`
		Value     string `json:"value"`
		Temporary bool   `json:"temporary"`
	}
	tempPasswordChange := &PayloadPasswordChange{Type: "password", Value: newPassword, Temporary: false}
	jsonPasswordChange, err := json.Marshal(tempPasswordChange)
	payload := strings.NewReader(string(jsonPasswordChange))
	req, err := http.NewRequest("PUT", url, payload)
	if err != nil {
		return &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}

	req.Header.Add("Authorization", "Bearer "+accesstoken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}

	defer res.Body.Close()

	// handle HTTP status codes
	if res.StatusCode != http.StatusNoContent {
		errNotFound := errors.New(res.Status)
		return &cwerrors.BasicError{errOpNotFound, errNotFound, errNotFound.Error()}
	}

	return nil
}

// SecUserAddRole : Adds a role to a specified user
func SecUserAddRole(c *cli.Context) *cwerrors.BasicError {
	hostname := strings.TrimSpace(strings.ToLower(c.String("host")))
	realm := strings.TrimSpace(c.String("realm"))
	accesstoken := strings.TrimSpace(c.String("accesstoken"))
	targetUser := strings.TrimSpace(c.String("name"))
	roleName := strings.TrimSpace(c.String("role"))

	// lookup an existing user
	logr.Tracef("Looking up user : %v", targetUser)
	registeredUser, secErr := SecUserGet(c)
	if secErr != nil {
		return secErr
	}

	// get the existing role
	existingRole, secErr := getRoleByName(c, roleName)
	if secErr != nil {
		return secErr
	}

	// build REST request
	logr.Printf("Adding role '%v' to user : '%v'", existingRole.Name, registeredUser.ID)
	url := hostname + "/auth/admin/realms/" + realm + "/users/" + registeredUser.ID + "/role-mappings/realm"

	type PayloadRole struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	listOfRoles := []PayloadRole{{ID: existingRole.ID, Name: existingRole.Name}}
	jsonRolesToAdd, err := json.Marshal(listOfRoles)
	payload := strings.NewReader(string(jsonRolesToAdd))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}

	req.Header.Add("Authorization", "Bearer "+accesstoken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &cwerrors.BasicError{errOpConnection, err, err.Error()}
	}

	// handle HTTP status codes (success returns status code StatusNoContent)
	if res.StatusCode != http.StatusNoContent {
		errNotFound := errors.New(res.Status)
		return &cwerrors.BasicError{errOpNotFound, errNotFound, errNotFound.Error()}
	}

	return nil
}
