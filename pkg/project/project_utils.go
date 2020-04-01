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
type (
	ProjectError struct {
		Op   string
		Err  error
		Desc string
	}
)

const (
	errBadPath           = "proj_path" // Invalid path provided
	errBadType           = "proj_type" // Invalid type provided
	errOpBind            = "proj_bind"
	errOpRequest         = "proj_request"
	errOpResponse        = "proj_response" // Bad response to http
	errOpFileParse       = "proj_parse"
	errOpFileLoad        = "proj_load"
	errOpFileWrite       = "proj_write"
	errOpFileDelete      = "proj_delete"
	errOpUnbind          = "proj_unbind"
	errOpGetProject      = "proj_get"
	errOpCreateProject   = "project create"
	errOpConflict        = "proj_conflict"
	errOpNotFound        = "proj_notfound"
	errOpConNotFound     = "connection_notfound"
	errOpInvalidID       = "proj_id_invalid"
	errOpInvalidOptions  = "proj_options_invalid"
	errOpSync            = "proj_sync"
	errOpSyncRef         = "proj_sync_ref"
	errOpWriteCwSettings = "proj_write_cw_settings"
)

const (
	textDupName                   = "project name is already in use"
	textInvalidType               = "project type is invalid"
	textInvalidProjectID          = "project ID is invalid"
	textConnectionExists          = "project already added to this connection"
	textConMissing                = "project connection not found"
	textNoCodewind                = "unable to connect to Codewind server"
	textAPINotFound               = "unable to find requested resource on Codewind server"
	textNoProjects                = "unable to find any codewind projects"
	textUpgradeError              = "error occurred upgrading projects"
	textNoProjectPath             = "project path not given"
	textProjectPathDoesNotExist   = "given project path does not exist"
	textProjectPathNonEmpty       = "Non empty directory provided"
	textUnknownResponseCode       = "unknown response code returned from Codewind server"
	textProjectLinkTargetNotFound = "target project not found on Codewind server"
	textProjectLinkConflict       = "project link env is already in use"
	textInvalidRequest            = "request parameters are invalid"
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
