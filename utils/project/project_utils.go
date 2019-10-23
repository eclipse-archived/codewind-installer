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
	"regexp"
)

// ProjectError : A Project error
type ProjectError struct {
	Op   string
	Err  error
	Desc string
}

const (
	errBadPath     = "proj_path"     // Invalid path provided
	errBadType     = "proj_type"     // Invalid type provided
	errOpResponse  = "proj_response" // Bad response to http
	errOpFileParse = "proj_parse"
	errOpFileLoad  = "proj_load"
	errOpFileWrite = "proj_write"
	errOpConflict  = "proj_conflict"
	errOpNotFound  = "proj_notfound"
	errOpInvalidID = "proj_id_invalid"
)

const (
	textDupName          = "project name is already in use"
	textInvalidType      = "project type is invalid"
	textInvalidProjectID = "project ID is invalid"
	textDeploymentExists = "project already added to this deployment"
	textDepMissing       = "project deployment not found"
)

// ProjectError : Error formatted in JSON containing an errorOp and a description from
// either a fault condition in the CLI, or an error payload from a REST request
func (pe *ProjectError) Error() string {
	type Output struct {
		Operation   string `json:"error"`
		Description string `json:"error_description"`
	}
	tempOutput := &Output{Operation: pe.Op, Description: pe.Err.Error()}
	jsonError, _ := json.Marshal(tempOutput)
	return string(jsonError)
}

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
