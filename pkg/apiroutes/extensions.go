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

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

// GetExtensions gets project extensions from PFE's REST API.
func GetExtensions(conID string) ([]utils.Extension, error) {
	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, conInfoErr.Err
	}
	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return nil, conErr.Err
	}

	req, err := http.NewRequest("GET", conURL+"/api/v1/extensions", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, httpSecError := sechttp.DispatchHTTPRequest(client, req, conInfo)
	if httpSecError != nil {
		return nil, httpSecError
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var extensions []utils.Extension
	err = json.Unmarshal(byteArray, &extensions)
	if err != nil {
		return nil, err
	}

	return extensions, nil
}
