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
	"errors"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
)

/// GetConnectionID : Gets the the connectionID for a given projectID
func GetConnectionID(projectID string) (string, *ProjectError) {
	allConnections, getConConfigErr := connections.GetConnectionsConfig()
	if getConConfigErr != nil {
		return "", &ProjectError{errOpConNotFound, getConConfigErr, getConConfigErr.Error()}
	}

	for i := 0; i < len(allConnections.Connections); i++ {
		currentConID := allConnections.Connections[i].ID

		conInfo, conInfoErr := connections.GetConnectionByID(currentConID)
		if conInfoErr != nil {
			return "", &ProjectError{errOpConNotFound, conInfoErr, conInfoErr.Error()}
		}

		conURL, conErr := config.PFEOriginFromConnection(conInfo)
		if conErr != nil {
			// Skip the connection if it's not running (local will error here)
			continue
		}

		projects, getAllErr := GetAll(http.DefaultClient, conInfo, conURL)
		if getAllErr != nil {
			// Skip the connection if it's not running (remote will error here)
			continue
		}

		for _, project := range projects {
			if project.ProjectID == projectID {
				return currentConID, nil
			}
		}
	}
	// We haven't found the project on any active connection so return an error
	projError := errors.New("Active connection not found for project " + projectID)
	return "", &ProjectError{errOpConNotFound, projError, projError.Error()}
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
