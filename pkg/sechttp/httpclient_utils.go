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

package sechttp

import "encoding/json"

// HTTPSecError : secclient package errors
type HTTPSecError struct {
	Op   string
	Err  error
	Desc string
}

const (
	errOpNoConnection = "tx_connection"
	errOpAuthFailed   = "tx_auth"
	errOpFailed       = "tx_failed"
	errOpNoPassword   = "tx_nopassword"
)

const (
	errConnetionNotFound = "Cant find a valid connection"
	errMissingPassword   = "Unable to find password in keychain"
)

// HTTPSecError : Error formatted in JSON containing an errorOp and a description from
// either a fault condition in the CLI, or an error payload from a REST request
func (se *HTTPSecError) Error() string {
	type Output struct {
		Operation   string `json:"error"`
		Description string `json:"error_description"`
	}
	tempOutput := &Output{Operation: se.Op, Description: se.Err.Error()}
	jsonError, _ := json.Marshal(tempOutput)
	return string(jsonError)
}
