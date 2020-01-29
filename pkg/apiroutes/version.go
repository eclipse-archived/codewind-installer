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

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type (
	// ContainerVersionsList : sdf
	ContainerVersionsList struct {
		CwctlVersion string                       `json:"cwctlVersion"`
		Connections  map[string]ContainerVersions `json:"connections"`
	}

	// ContainerVersions : The versions of the Codewind containers that are running
	ContainerVersions struct {
		CwctlVersion       string `json:"cwctlVersion,omitempty"`
		PerformanceVersion string `json:"performanceVersion"`
		GatekeeperVersion  string `json:"gatekeeperVersion,omitempty"`
		PFEVersion         string `json:"PFEVersion"`
	}

	// EnvResponse : The relevant response fields from the remote environment API
	EnvResponse struct {
		Version        string `json:"codewind_version"`
		ImageBuildTime string `json:"image_build_time"`
	}
)

// GetAllContainerVersions : Get the versions of each Codewind container for each given connection ID
func GetAllContainerVersions(conIDList []string, cwctlVersion string, httpClient utils.HTTPClient) (ContainerVersionsList, error) {
	var containerVersionsList ContainerVersionsList
	containerVersionsList.CwctlVersion = cwctlVersion

	var connectionVersions = make(map[string]ContainerVersions)
	for _, conID := range conIDList {
		containerVersion, containerVersionErr := GetContainerVersions(conID, "", httpClient)
		if containerVersionErr != nil {
			return ContainerVersionsList{}, containerVersionErr
		}
		connectionVersions[conID] = containerVersion
	}
	containerVersionsList.Connections = connectionVersions

	return containerVersionsList, nil
}

// GetContainerVersions : Get the versions of each Codewind container, for a given connection ID
func GetContainerVersions(conID, cwctlVersion string, httpClient utils.HTTPClient) (ContainerVersions, error) {
	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return ContainerVersions{}, conInfoErr.Err
	}

	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return ContainerVersions{}, conErr.Err
	}

	var containerVersions ContainerVersions
	PFEVersion, err := GetPFEVersionFromConnection(conInfo, conURL, http.DefaultClient)
	if err != nil {
		return ContainerVersions{}, err
	}

	PerformanceVersion, err := GetPerformanceVersionFromConnection(conInfo, conURL, http.DefaultClient)
	if err != nil {
		return ContainerVersions{}, err
	}

	// Add cwctlVersion if it is passed in
	if cwctlVersion != "" {
		containerVersions.CwctlVersion = cwctlVersion
	}

	containerVersions.PFEVersion = PFEVersion
	containerVersions.PerformanceVersion = PerformanceVersion

	if conID != "local" {
		GatekeeperVersion, err := GetGatekeeperVersionFromConnection(conInfo, http.DefaultClient)
		if err != nil {
			return ContainerVersions{}, err
		}

		containerVersions.GatekeeperVersion = GatekeeperVersion
	}

	return containerVersions, nil
}

// GetPFEVersionFromConnection : Get the version of the PFE container, deployed to the connection with the given ID
func GetPFEVersionFromConnection(connection *connections.Connection, url string, HTTPClient utils.HTTPClient) (string, error) {
	req, err := http.NewRequest("GET", url+"/api/v1/environment", nil)
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
func GetPerformanceVersionFromConnection(connection *connections.Connection, url string, HTTPClient utils.HTTPClient) (string, error) {
	req, err := http.NewRequest("GET", url+"/performance/api/v1/environment", nil)
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
	resp, httpSecError := sechttp.DispatchHTTPRequest(HTTPClient, req, connection)
	if httpSecError != nil {
		return "", httpSecError
	}
	// Set version field to empty string, if API call not successful
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

	if env.ImageBuildTime != "" {
		codewindVersion = codewindVersion + "-" + env.ImageBuildTime
	}

	return codewindVersion, nil
}
