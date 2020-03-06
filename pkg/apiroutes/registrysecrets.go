/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

package apiroutes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type (
	// RegistryResponse : The registry information
	RegistryResponse struct {
		Address  string `json:"address"`
		Username string `json:"username"`
	}

	// RegistryParameters : The request structure to set the log level
	RegistryParameters struct {
		Address     string `json:"address"`
		Credentials string `json:"credentials"`
	}

	// Credentials : The registry credentials, sent as a base64 encoded string.
	Credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// AddressParameter : Used when removing credentials.
	AddressParameter struct {
		Address string `json:"address"`
	}
)

// GetRegistrySecrets : Get the current registry secrets for the PFE container
func GetRegistrySecrets(conInfo *connections.Connection, conURL string, httpClient utils.HTTPClient) (*[]RegistryResponse, error) {
	req, err := http.NewRequest("GET", conURL+"/api/v1/registrysecrets", nil)
	if err != nil {
		return nil, err
	}

	return handleRegistrySecretsResponse(req, conInfo, httpClient, http.StatusOK)
}

// AddRegistrySecret : Set a registry secret in the PFE container
func AddRegistrySecret(conInfo *connections.Connection, conURL string, httpClient utils.HTTPClient, address string, username string, password string) (*[]RegistryResponse, error) {

	// The username and password are sent inside a base64 encoded field in the jsonPayload.
	credentials := &Credentials{Username: username, Password: password}
	credentialsStr, _ := json.Marshal(credentials)
	credentialsBase64 := base64.StdEncoding.EncodeToString([]byte(credentialsStr))
	registryParameters := &RegistryParameters{Address: address, Credentials: credentialsBase64}
	jsonPayload, _ := json.Marshal(registryParameters)

	req, err := http.NewRequest("POST", conURL+"/api/v1/registrysecrets", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return handleRegistrySecretsResponse(req, conInfo, httpClient, http.StatusCreated)
}

// RemoveRegistrySecret : Remove a registry secret from the PFE container
func RemoveRegistrySecret(conInfo *connections.Connection, conURL string, httpClient utils.HTTPClient, address string) (*[]RegistryResponse, error) {

	// The username and password are sent inside a base64 encoded field in the jsonPayload.
	addressParameter := &AddressParameter{Address: address}
	jsonPayload, _ := json.Marshal(addressParameter)

	req, err := http.NewRequest("DELETE", conURL+"/api/v1/registrysecrets", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return handleRegistrySecretsResponse(req, conInfo, httpClient, http.StatusOK)
}

// All three API calls (GET, POST and DELETE) return the same response.
func handleRegistrySecretsResponse(req *http.Request, conInfo *connections.Connection, httpClient utils.HTTPClient, successCode int) (*[]RegistryResponse, error) {
	resp, httpSecError := sechttp.DispatchHTTPRequest(httpClient, req, conInfo)
	if httpSecError != nil {
		return nil, httpSecError
	}

	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != successCode {
		return nil, fmt.Errorf("%s - %s", http.StatusText(resp.StatusCode), string(byteArray))
	}

	var registrySecrets []RegistryResponse
	err = json.Unmarshal(byteArray, &registrySecrets)
	if err != nil {
		return nil, err
	}

	return &registrySecrets, nil
}
