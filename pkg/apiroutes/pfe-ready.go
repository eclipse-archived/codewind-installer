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
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/utils"
)

// IsPFEReady : Get PFE Ready for connection
func IsPFEReady(httpClient utils.HTTPClient, host string) (bool, error) {
	req, err := http.NewRequest("GET", host+"/ready", nil)
	if err != nil {
		return false, err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	status := res.StatusCode
	if status == 200 {
		return true, nil
	}
	return false, nil
}
