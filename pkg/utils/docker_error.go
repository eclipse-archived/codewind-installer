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

const (
	errOpValidate            = "DOCKER_VALIDATE"
	errOpClientCreate        = "CLIENT_CREATE_ERROR"
	errOpContainerInspect    = "CONTAINER_INSPECT_ERROR"
	errOpContainerError      = "CONTAINER_ERROR"
	errOpStopContainer       = "CONTAINER_STOP_ERROR"
	errOpDockerComposeStart  = "DOCKER_COMPOSE_START_ERROR"
	errOpDockerComposeStop   = "DOCKER_COMPOSE_STOP_ERROR"
	errOpDockerComposeRemove = "DOCKER_COMPOSE_REMOVE"
	errOpImageNotFound       = "IMAGE_NOT_FOUND"
	errOpImagePull           = "IMAGE_PULL_ERROR"
	errOpImageTag            = "IMAGE_TAG_ERROR"
	errOpImageRemove         = "IMAGE_REMOVE_ERROR"
	errOpImageDigest         = "IMAGE_DIGEST_ERROR"
	errOpContainerList       = "CONTAINER_LIST_ERROR"
	errOpImageList           = "IMAGE_LIST_ERROR"
)

const (
	textBadDigest = "Failed to validate docker image checksum"
)
