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

package utils

import "encoding/json"

// DockerError struct will format the error
type DockerError struct {
	Op   string
	Err  error
	Desc string
}

const (
	errOpValidate = "docker_validate" // validate docker images
)

const (
	textBadDigest = "Failed to validate docker image checksum"
)

// DockerError : Error formatted in JSON containing an errorOp and a description
func (de *DockerError) Error() string {
	type Output struct {
		Operation   string `json:"error"`
		Description string `json:"error_description"`
	}
	tempOutput := &Output{Operation: de.Op, Description: de.Err.Error()}
	jsonError, _ := json.Marshal(tempOutput)
	return string(jsonError)
}
