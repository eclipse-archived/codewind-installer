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

package apiroutes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/appconstants"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type (
	// ContainerVersions : The versions of the Codewind containers that are running
	ContainerVersions struct {
		CwctlVersion       string
		PerformanceVersion string
		GatekeeperVersion  string
		PFEVersion         string
	}

	// EnvResponse : The response from the remote environment API
	EnvResponse struct {
		Version        string `json:"codewind_version"`
		ImageBuildTime string `json:"image_build_time"`
	}
)

// GetContainerVersions : Get the versions of each Codewind container, for a given connection ID
func GetContainerVersions(conID string, httpClient utils.HTTPClient) (ContainerVersions, error) {
	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return ContainerVersions{}, conInfoErr.Err
	}

	var containerVersions ContainerVersions
	PFEVersion, err := GetPFEVersionFromConnection(conInfo, http.DefaultClient)
	if err != nil {
		return ContainerVersions{}, err
	}

	GatekeeperVersion, err := GetGatekeeperVersionFromConnection(conInfo, http.DefaultClient)
	if err != nil {
		return ContainerVersions{}, err
	}

	PerformanceVersion, err := GetPerformanceVersionFromConnection(conInfo, http.DefaultClient)
	if err != nil {
		return ContainerVersions{}, err
	}

	containerVersions.CwctlVersion = appconstants.VersionNum
	containerVersions.PFEVersion = PFEVersion
	containerVersions.GatekeeperVersion = GatekeeperVersion
	containerVersions.PerformanceVersion = PerformanceVersion

	return containerVersions, nil
}

// GetPFEVersionFromConnection : Get the version of the PFE container, deployed to the connection with the given ID
func GetPFEVersionFromConnection(connection *connections.Connection, HTTPClient utils.HTTPClient) (string, error) {
	req, err := http.NewRequest("GET", connection.URL+"/api/v1/environmen", nil)
	if err != nil {
		return "", err
	}

	version, err := getVersionFromEnvAPI(req, connection, HTTPClient)
	if err != nil {
		return "", err
	}
	return version, err
}

// GetGatekeeperVersionFromConnection : Get the version of the Gatekeeper container, deployed to the connection with the given ID
func GetGatekeeperVersionFromConnection(connection *connections.Connection, HTTPClient utils.HTTPClient) (string, error) {
	req, err := http.NewRequest("GET", connection.URL+"/api/v1/gatekeeper/environment", nil)
	if err != nil {
		return "", err
	}

	version, err := getVersionFromEnvAPI(req, connection, HTTPClient)
	if err != nil {
		return "", err
	}
	return version, err
}

// GetPerformanceVersionFromConnection : Get the version of the Performance container, deployed to the connection with the given ID
func GetPerformanceVersionFromConnection(connection *connections.Connection, HTTPClient utils.HTTPClient) (string, error) {
	req, err := http.NewRequest("GET", connection.URL+"/performance/api/v1/environment", nil)
	if err != nil {
		return "", err
	}

	version, err := getVersionFromEnvAPI(req, connection, HTTPClient)
	if err != nil {
		return "", err
	}
	return version, err
}

func getVersionFromEnvAPI(req *http.Request, connection *connections.Connection, HTTPClient utils.HTTPClient) (string, error) {
	client := &http.Client{}
	resp, httpSecError := sechttp.DispatchHTTPRequest(client, req, connection)
	if httpSecError != nil {
		return "", httpSecError
	}
	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var env EnvResponse
	err = json.Unmarshal(byteArray, &env)
	if err != nil {
		return "", err
	}
	codewindVersion := env.Version

	if env.ImageBuildTime != "nil" {
		codewindVersion = codewindVersion + "-" + env.ImageBuildTime
	}

	return codewindVersion, nil
}
