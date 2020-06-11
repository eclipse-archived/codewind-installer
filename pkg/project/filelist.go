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

type FileList []string

// GetProjectFromID : Get project details from Codewind
func GetProjectFileList(httpClient utils.HTTPClient, connection *connections.Connection, url, projectID string) (FileList, *ProjectError) {
	req, requestErr := http.NewRequest("GET", url+"/api/v1/projects/"+projectID+"/fileList", nil)
	if requestErr != nil {
		return nil, &ProjectError{errOpRequest, requestErr, requestErr.Error()}
	}

	// send request
	resp, httpSecError := sechttp.DispatchHTTPRequest(httpClient, req, connection)
	if httpSecError != nil {
		return nil, &ProjectError{errOpRequest, httpSecError, httpSecError.Desc}
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respErr := errors.New(textAPINotFound)
		return nil, &ProjectError{errOpNotFound, respErr, textAPINotFound}
	}

	byteArray, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, &ProjectError{errOpRequest, readErr, readErr.Error()}
	}

	var list FileList
	jsonErr := json.Unmarshal(byteArray, &list)
	if jsonErr != nil {
		return nil, &ProjectError{errOpRequest, jsonErr, jsonErr.Error()}
	}
	return list, nil
}
