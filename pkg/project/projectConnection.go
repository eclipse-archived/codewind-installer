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

	"github.com/eclipse/codewind-installer/pkg/connections"
)

// ConnectionFile : Structure of the project-connections file
type ConnectionFile struct {
	SchemaVersion int    `json:"schemaVersion"`
	ID            string `json:"connectionID"`
}

const connectionTargetSchemaVersion = 1

// SetConnection : Add a connection target
func SetConnection(conID string, projectID string) *ProjectError {

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

	connectionTargets, projError := loadConnectionFile(projectID)
	if projError != nil && connectionTargets == nil {
		err := CreateConnectionFile(projectID)
		if err != nil {
			return err
		}
	}

	connectionTargets, _ = loadConnectionFile(projectID)
	connectionTargets.ID = conID

	// Save the project-connections file
	projError = saveConnectionTargets(projectID, connectionTargets)
	if projError != nil {
		return projError
	}
	return nil
}

// ResetConnectionFile : Reset target file
func ResetConnectionFile(projectID string) *ProjectError {
	connectionTargets := ConnectionFile{
		SchemaVersion: connectionTargetSchemaVersion,
		ID:            "local",
	}
	projError := saveConnectionTargets(projectID, &connectionTargets)
	if projError != nil {
		return projError
	}
	return nil
}

// GetConnection : List the connection for a projectID
func GetConnection(projectID string) (*ConnectionFile, *ProjectError) {
	connectionTargets, projErr := loadConnectionFile(projectID)
	if projErr != nil {
		return nil, projErr
	}
	return connectionTargets, nil
}

// GetConnectionID : Gets the the connectionID for a given projectID
func GetConnectionID(projectID string) (string, *ProjectError) {
	connection, err := GetConnection(projectID)
	if err != nil {
		return "", err
	}
	var conID string
	if connection.ID != "" {
		conID = connection.ID
	} else {
		projError := errors.New("Connection not found for project " + projectID)
		return "", &ProjectError{errOpConNotFound, projError, projError.Error()}
	}
	return conID, nil
}

// ConnectionFileExists : Returns true if connection file exists for the projectID
func ConnectionFileExists(projectID string) bool {
	info, err := os.Stat(getConnectionFilename(projectID))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// CreateConnectionFile : Creates the /connections/{project_id}.json file if one doesn't exist, with default local connection
func CreateConnectionFile(projectID string) *ProjectError {
	_, err := os.Stat(getConnectionFilename(projectID))
	if os.IsNotExist(err) {
		os.MkdirAll(getProjectConnectionConfigDir(), 0777)
		resetFileError := ResetConnectionFile(projectID)
		if resetFileError != nil {
			return resetFileError
		}
	}
	return nil
}

// RemoveConnectionFile : Remove the connection file for a project
func RemoveConnectionFile(projectID string) *ProjectError {
	// delete file
	var err = os.Remove(getConnectionFilename(projectID))
	if err != nil {
		return &ProjectError{errOpFileDelete, err, err.Error()}
	}
	return nil
}

// getProjectConnectionConfigDir : Get directory path to the connection file
func getProjectConnectionConfigDir() string {
	val, isSet := os.LookupEnv("CHE_API_EXTERNAL")
	homeDir := ""
	if isSet && (val != "") {
		val, isSet := os.LookupEnv("CHE_PROJECTS_ROOT")
		if isSet && (val != "") {
			homeDir = val
		} else {
			// Cannot set projects root without env variable, suggests issue with Codewind Che installation
			panic("CHE_PROJECTS_ROOT not set")
		}
	} else {
		const GOOS string = runtime.GOOS
		if GOOS == "windows" {
			homeDir = os.Getenv("USERPROFILE")
		} else {
			homeDir = os.Getenv("HOME")
		}
	}
	return path.Join(homeDir, ".codewind", "config", "connections")
}

// getConnectionFilename : Get full file path of connection file
func getConnectionFilename(projectID string) string {
	return path.Join(getProjectConnectionConfigDir(), projectID+".json")
}

// saveConnectionTargets : Write the targets file in JSON format
func saveConnectionTargets(projectID string, connectionTargets *ConnectionFile) *ProjectError {
	body, err := json.MarshalIndent(connectionTargets, "", "\t")
	if err != nil {
		return &ProjectError{errOpFileParse, err, err.Error()}
	}
	projError := ioutil.WriteFile(getConnectionFilename(projectID), body, 0644)
	if projError != nil {
		return &ProjectError{errOpFileWrite, projError, projError.Error()}
	}
	return nil
}

// loadConnectionFile :  Loads the connection file for a project
func loadConnectionFile(projectID string) (*ConnectionFile, *ProjectError) {
	projectID = strings.ToLower(projectID)
	filePath := getConnectionFilename(projectID)
	file, err := ioutil.ReadFile(filePath)

	if err != nil {
		return nil, &ProjectError{errOpFileLoad, err, err.Error()}
	}

	// parse the file
	projectConnectionTargets := ConnectionFile{}
	err = json.Unmarshal([]byte(file), &projectConnectionTargets)
	if err != nil {
		return nil, &ProjectError{errOpFileParse, err, err.Error()}
	}
	return &projectConnectionTargets, nil
}
