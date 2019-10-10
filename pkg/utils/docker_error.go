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
	errOpValidate          = "DOCKER_VALIDATE"
	errOpClientCreate      = "CLIENT_CREATE_ERROR"
	errOpImageNotFound     = "IMAGE_NOT_FOUND"
	errOpImagePull         = "IMAGE_PULL_ERROR"
	errOpContainersStopped = "CONTAINERS_STOPPED"
	errOpContainerError    = "CONTAINER_ERROR"
	errOpDockerCompose     = "DOCKER_COMPOSE_ERROR"
	errOpImageTag          = "IMAGE_TAG_ERROR"
	errOpImageRemove       = "IMAGE_REMOVE_ERROR"
	errOpContainerList     = "CONTAINER_LIST_ERROR"
	errOpImageList         = "IMAGE_LIST_ERROR"
	errOpNetworkList       = "NETWORK_LIST_ERROR"
	errOpNetworkError      = "DOCKER_NETWORK_ERROR"
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
