package project

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type (
	// Project : Represents a project
	Project struct {
		ProjectID      string `json:"projectID"`
		Name           string `json:"name"`
		Language       string `json:"language"`
		Host           string `json:"host"`
		LocationOnDisk string `json:"locOnDisk"`
		AppStatus      string `json:"appStatus"`
	}
)

// GetProjectFromID : Get project details from Codewind
func GetProjectFromID(httpClient utils.HTTPClient, connection *connections.Connection, url, projectID string) (*Project, *ProjectError) {
	req, requestErr := http.NewRequest("GET", url+"/api/v1/projects/"+projectID+"/", nil)
	if requestErr != nil {
		return nil, &ProjectError{errOpRequest, requestErr, requestErr.Error()}
	}

	// send request
	resp, httpSecError := sechttp.DispatchHTTPRequest(httpClient, req, connection)
	if httpSecError != nil {
		return nil, &ProjectError{errOpRequest, httpSecError, httpSecError.Desc}
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		respErr := errors.New(textAPINotFound)
		return nil, &ProjectError{errOpNotFound, respErr, textAPINotFound}
	}

	byteArray, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, &ProjectError{errOpRequest, readErr, readErr.Error()}
	}
	var project Project
	jsonErr := json.Unmarshal(byteArray, &project)
	if jsonErr != nil {
		return nil, &ProjectError{errOpRequest, jsonErr, jsonErr.Error()}
	}
	return &project, nil
}

// GetProjectIDFromName : Get a project ID using its name
func GetProjectIDFromName(httpClient utils.HTTPClient, connection *connections.Connection, url, projectName string) (string, *ProjectError) {
	projects, err := GetAll(httpClient, connection, url)
	if err != nil {
		return "", err
	}

	for _, project := range projects {
		if project.Name == projectName {
			return project.ProjectID, nil
		}
	}
	respErr := errors.New(textAPINotFound)
	return "", &ProjectError{errOpNotFound, respErr, textAPINotFound}
}

// GetProjectFromName : Get a project using its name
func GetProjectFromName(httpClient utils.HTTPClient, connection *connections.Connection, url, projectName string) (*Project, *ProjectError) {
	projectID, getProjectIDError := GetProjectIDFromName(httpClient, connection, url, projectName)
	if getProjectIDError != nil {
		return nil, getProjectIDError
	}

	project, getProjectError := GetProjectFromID(httpClient, connection, url, projectID)
	if getProjectError != nil {
		return nil, getProjectError
	}

	return project, nil
}
