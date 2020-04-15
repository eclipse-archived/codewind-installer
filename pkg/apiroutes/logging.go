/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

package apiroutes

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type (
	// LoggingResponse : The logging level information
	LoggingResponse struct {
		CurrentLevel string   `json:"currentLevel"`
		DefaultLevel string   `json:"defaultLevel"`
		AllLevels    []string `json:"allLevels"`
	}

	// LogParameter : The request structure to set the log level
	LogParameter struct {
		Level string `json:"level"`
	}
)

// GetLogLevel : Get the current log level for the PFE container
func GetLogLevel(conInfo *connections.Connection, conURL string, httpClient utils.HTTPClient) (LoggingResponse, error) {
	req, err := http.NewRequest("GET", conURL+"/api/v1/logging", nil)
	if err != nil {
		return LoggingResponse{}, err
	}

	resp, httpSecError := sechttp.DispatchHTTPRequest(httpClient, req, conInfo)
	if httpSecError != nil {
		return LoggingResponse{}, httpSecError
	}

	if resp.StatusCode != http.StatusOK {
		return LoggingResponse{}, errors.New(http.StatusText(resp.StatusCode))
	}

	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return LoggingResponse{}, err
	}

	var loggingLevels LoggingResponse
	err = json.Unmarshal(byteArray, &loggingLevels)
	if err != nil {
		return LoggingResponse{}, err
	}

	return loggingLevels, nil
}

// SetLogLevel : Set the current log level for the PFE container
func SetLogLevel(conInfo *connections.Connection, conURL string, httpClient utils.HTTPClient, newLogLevel string) error {
	// Send the new logging level.
	logLevel := &LogParameter{Level: newLogLevel}
	jsonPayload, _ := json.Marshal(logLevel)
	req, err := http.NewRequest("PUT", conURL+"/api/v1/logging", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, httpSecError := sechttp.DispatchHTTPRequest(httpClient, req, conInfo)
	if httpSecError != nil {
		return httpSecError
	}

	if resp.StatusCode == 400 {
		return errors.New("Invalid log level")
	} else if resp.StatusCode == 500 {
		return errors.New("Error setting log level")
	} else if resp.StatusCode != http.StatusOK {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return nil
}
