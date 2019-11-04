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

package apiroutes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/eclipse/codewind-installer/config"
	"github.com/eclipse/codewind-installer/utils"
)

type IgnoredFiles []string

// GetIgnoredPaths calls pfe to get the default ignoredPaths for that projectType
func GetIgnoredPaths(httpClient utils.HTTPClient, projectType string) (IgnoredFiles, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", config.PFEOrigin()+"/api/v1/ignoredPaths", nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("projectType", projectType)

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ignoredFiles IgnoredFiles
	err = json.Unmarshal(byteArray, &ignoredFiles)
	if err != nil {
		return nil, err
	}
	return ignoredFiles, nil
}
