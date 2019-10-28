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

package project

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/eclipse/codewind-installer/config"

	"github.com/eclipse/codewind-installer/utils/connections"
)

// Target : A Connection target
type Target struct {
	ConnectionID string `json:"id"`
}

// ConnectionTargets : Structure of the project-connections file
type ConnectionTargets struct {
	SchemaVersion     int    `json:"schemaVersion"`
	ConnectionTargets Target `json:"connectionTargets"`
}

const connectionTargetSchemaVersion = 1

// SetConnection : Add a connection target
func SetConnection(projectID string, conID string) *ProjectError {

	connection, conErr := connections.GetConnectionByID(conID)
	if conErr != nil || connection == nil {
		projError := errors.New("Connection unknown")
		return &ProjectError{"con_not_found", projError, projError.Error()}
	}

	// Check if projectID is supplied in correct format
	if !IsProjectIDValid(projectID) {
		projError := errors.New(textInvalidProjectID)
		return &ProjectError{errOpInvalidID, projError, projError.Error()}
	}

	// Load the project-connection.json
	connectionTargets, projError := loadTargets(projectID)

	if projError != nil && connectionTargets == nil {
		_, err := os.Stat(getProjectConnectionsFilename(projectID))
		if os.IsNotExist(err) {
			os.MkdirAll(getProjectConnectionConfigDir(), 0777)
			projErr := ResetTargetFile(projectID)
			if projErr != nil {
				return projErr
			}
		}
	}

	// Add the connection to the project-connections file
	target := Target{
		ConnectionID: conID,
	}
	connectionTargets.ConnectionTargets = target

	// Save the project-connections file
	projError = saveConnectionTargets(projectID, connectionTargets)
	if projError != nil {
		return projError
	}
	return nil
}

// ResetTargetFile : Reset target file
func ResetTargetFile(projectID string) *ProjectError {
	connectionTargets := ConnectionTargets{
		SchemaVersion: connectionTargetSchemaVersion,
	}
	projError := saveConnectionTargets(projectID, &connectionTargets)
	if projError != nil {
		return projError
	}
	return nil
}

// GetConnection : List the connection for a projectID
func GetConnection(projectID string) (*ConnectionTargets, *ProjectError) {
	connectionTargets, projErr := loadTargets(projectID)
	if projErr != nil {
		return nil, projErr
	}
	return connectionTargets, nil
}

// GetConnectionURL returns to the connection URL for a given projectID, unique to each project connection
func GetConnectionURL(projectID string) (string, *ProjectError) {
	conID, err := GetConnectionID(projectID)

	if err != nil {
		return "", err
	}

	projectConInfo, conErr := connections.GetConnectionByID(conID)
	if conErr != nil {
		return "", &ProjectError{errOpNotFound, conErr, conErr.Error()}
	}

	if conID == "local" {
		return config.PFEApiRoute(), nil
	}
	return projectConInfo.URL, nil
}

// GetConnectionID gets the the connectionID for a given projectID
func GetConnectionID(projectID string) (string, *ProjectError) {
	targetConnections, err := GetConnection(projectID)
	if err != nil {
		return "", err
	}
	conTargets := targetConnections.ConnectionTargets
	var conID string
	if conTargets.ConnectionID != "" {
		conID = conTargets.ConnectionID
	} else {
		projError := errors.New("Connection not found for project " + projectID)
		return "", &ProjectError{errOpConNotFound, projError, projError.Error()}
	}
	return conID, nil
}

// getProjectConnectionConfigDir : get directory path to the connections file
func getProjectConnectionConfigDir() string {
	const GOOS string = runtime.GOOS
	homeDir := ""
	if GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}
	return path.Join(homeDir, ".codewind", "config", "connections")
}

// getProjectConnectionsFilename  : get full file path of connections file
func getProjectConnectionsFilename(projectID string) string {
	return path.Join(getProjectConnectionConfigDir(), projectID+".json")
}

// saveConnectionTargets : write the targets file in JSON format
func saveConnectionTargets(projectID string, connectionTargets *ConnectionTargets) *ProjectError {
	body, err := json.MarshalIndent(connectionTargets, "", "\t")
	if err != nil {
		return &ProjectError{errOpFileParse, err, err.Error()}
	}
	projError := ioutil.WriteFile(getProjectConnectionsFilename(projectID), body, 0644)
	if projError != nil {
		return &ProjectError{errOpFileWrite, projError, projError.Error()}
	}
	return nil
}

// loadTargets :  Loads the config file for a project
func loadTargets(projectID string) (*ConnectionTargets, *ProjectError) {
	projectID = strings.ToLower(projectID)
	file, err := ioutil.ReadFile(getProjectConnectionsFilename(projectID))
	if err != nil {
		return nil, &ProjectError{errOpFileLoad, err, err.Error()}
	}

	// parse the file
	projectConnectionTargets := ConnectionTargets{}
	err = json.Unmarshal([]byte(file), &projectConnectionTargets)
	if err != nil {
		return nil, &ProjectError{errOpFileParse, err, err.Error()}
	}
	return &projectConnectionTargets, nil
}
