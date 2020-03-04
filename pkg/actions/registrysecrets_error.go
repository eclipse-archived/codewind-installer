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

// RegistryError struct will format the error
type RegistryError struct {
	Op   string
	Err  error
	Desc string
}

const (
	errOpListRegistries = "LIST_REGISTRIES_ERROR"
	errOpAddRegistry    = "ADD_REGISTRY_ERROR"
	errOpRemoveRegistry = "DELETE_REGISTRY_ERROR"
)

// RegistryError : Error formatted in JSON containing an errorOp and a description
func (te *RegistryError) Error() string {
	type Output struct {
		Operation   string `json:"error"`
		Description string `json:"error_description"`
	}
	tempOutput := &Output{Operation: te.Op, Description: te.Err.Error()}
	jsonError, _ := json.Marshal(tempOutput)
	return string(jsonError)
}
