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

package actions

import "encoding/json"

// TemplateError struct will format the error
type TemplateError struct {
	Op   string
	Err  error
	Desc string
}

const (
	errOpListTemplates = "LIST_TEMPLATES_ERROR"
	errOpListStyles    = "LIST_STYLES_ERROR"
	errOpListRepos     = "LIST_REPOS_ERROR"
	errOpAddRepo       = "ADD_REPO_ERROR"
	errOpDeleteRepo    = "DELETE_REPO_ERROR"
	errOpEnableRepo    = "ENABLE_REPO_ERROR"
	errOpDisableRepo   = "DISABLE_REPO_ERROR"
)

// TemplateError : Error formatted in JSON containing an errorOp and a description
func (te *TemplateError) Error() string {
	type Output struct {
		Operation   string `json:"error"`
		Description string `json:"error_description"`
	}
	tempOutput := &Output{Operation: te.Op, Description: te.Err.Error()}
	jsonError, _ := json.Marshal(tempOutput)
	return string(jsonError)
}
