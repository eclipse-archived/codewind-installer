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
	"regexp"
)

const (
	errBadPath       = "proj_path"     // Invalid path provided
	errBadType       = "proj_type"     // Invalid type provided
	errOpResponse    = "proj_response" // Bad response to http
	errOpFileParse   = "proj_parse"
	errOpFileLoad    = "proj_load"
	errOpFileWrite   = "proj_write"
	errOpFileDelete  = "proj_delete"
	errOpGetProject  = "proj_get"
	errOpConflict    = "proj_conflict"
	errOpNotFound    = "proj_notfound"
	errOpConNotFound = "connection_notfound"
	errOpInvalidID   = "proj_id_invalid"
)

const (
	textDupName          = "project name is already in use"
	textInvalidType      = "project type is invalid"
	textInvalidProjectID = "project ID is invalid"
	textConnectionExists = "project already added to this connection"
	textConMissing       = "project connection not found"
	textNoCodewind       = "unable to connect to Codewind server"
	textAPINotFound      = "unable to find requested resource on Codewind server"
	textNoProjects       = "unable to find any codewind projects"
	textUpgradeError     = "error occurred upgrading projects"
)

// Result : status message
type Result struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
}

// IsProjectIDValid : Checks if a supplied project ID is in the correct format
func IsProjectIDValid(projectID string) bool {
	match, err := regexp.MatchString("[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}", projectID)
	if err != nil {
		return false
	}
	return match
}
