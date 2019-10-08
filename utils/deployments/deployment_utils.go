package deployments

import "encoding/json"

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

// DepError : Deployment package errors
type DepError struct {
	Op   string
	Err  error
	Desc string
}

const (
	// ErrOpSchema : Schema type errors
	ErrOpSchema = "sec_schema"

	// ErrOpTarget : Deployment target errors
	ErrOpTarget = "sec_target"
)

const (
	// TextTargetNotFound : missing target deployment text
	TextTargetNotFound = "Target deployment not found"
)

// DepError : Error formatted in JSON containing an errorOp and a description from
// either a fault condition in the CLI, or an error payload from a REST request
func (se *DepError) Error() string {
	type Output struct {
		Operation   string `json:"error"`
		Description string `json:"error_description"`
	}
	tempOutput := &Output{Operation: se.Op, Description: se.Err.Error()}
	jsonError, _ := json.Marshal(tempOutput)
	return string(jsonError)
}
