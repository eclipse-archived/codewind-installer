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

package project

import (
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

// Unbind a project from Codewind
func Unbind(httpClient utils.HTTPClient, conID, projectID string) error {
	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return conInfoErr.Err
	}
	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		return conErr.Err
	}
	req, err := http.NewRequest("POST", conURL+"/api/v1/projects/"+projectID+"/unbind", nil)
	if err != nil {
		return err
	}

	// send request
	res, httpSecError := sechttp.DispatchHTTPRequest(httpClient, req, conInfo)
	if httpSecError != nil {
		return httpSecError
	}

	defer res.Body.Close()
	return nil
}
