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

package errors

import (
	"errors"
	"log"
)

// CheckErr function to respond with appropriate error messages
func CheckErr(err error, code int, optMsg string) {
	if err != nil {
		switch code {
		case 100:
			log.Fatal("DOCKER_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 101:
			log.Fatal("DOCKER_COMPOSE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 102:
			log.Fatal("IMAGE_TAGGING_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 103:
			log.Fatal("CONTAINER_STATUS_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 104:
			log.Fatal("IMAGE_STATUS_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 105:
			log.Fatal("REMOVE_IMAGE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 107:
			log.Fatal("CONTAINER_LIST_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 108:
			log.Fatal("CONTAINER_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 109:
			log.Fatal("IMAGE_LIST_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 110:
			log.Fatal("DOCKER_NETWORK_LIST_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 111:
			log.Fatal("DOCKER_NETWORK_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 200:
			log.Fatal("INTERNAL_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 201:
			log.Fatal("CREATE_FILE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 202, 203, 204:
			log.Fatal("WRITE_FILE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 205:
			log.Fatal("DIRECTORY_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 206:
			// Do not want to exit if this is thrown
			log.Print("DELETE_FILE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 207:
			log.Fatal("READ_FILE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 208:
			log.Fatal("PARSING_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 300:
			log.Fatal("APPLICATION_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 400:
			log.Fatal("REPOSITORY_DOWNLOAD_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 401:
			log.Fatal("CREATE_ZIP_FILE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 402:
			log.Fatal("READ_ZIP_FILE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 403:
			log.Fatal("OUTPUT_FILE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		case 404:
			log.Fatal("WRITE_FILE_ERROR", "[", code, "]: ", err, ". ", optMsg)
		default:
			log.Fatal("UNKNOWN_ERROR", "[", code, "]: ", err, ". ", optMsg)
		}
	}
}

func newError(text string) error {
	return errors.New(text)
}
