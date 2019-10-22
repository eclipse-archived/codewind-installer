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
)

type ProjectError struct {
	Op   string
	Err  error
	Desc string
}

const (
	errBadPath = "proj_path" // Invalid path provided
	errBadType = "proj_type" // Invalid type provided
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
