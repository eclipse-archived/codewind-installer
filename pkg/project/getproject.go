package project

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type (
	// Project : Represents a project
	Project struct {
		ProjectID      string `json:"projectID"`
		LocationOnDisk string `json:"locOnDisk"`
	}
)

// GetProject : Get project details from Codewind
func GetProject(httpClient utils.HTTPClient, conID, projectID string) (*Project, error) {
	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, conInfoErr.Err
	}
	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return nil, conErr.Err
	}
	req, getProjectErr := http.NewRequest("GET", conURL+"/api/v1/projects/"+projectID+"/", nil)
	if getProjectErr != nil {
		fmt.Println(getProjectErr)
		return nil, getProjectErr
	}

	// send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	byteArray, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	var project Project
	getProjectErr = json.Unmarshal(byteArray, &project)
	if getProjectErr != nil {
		return nil, getProjectErr
	}
	return &project, nil
}
