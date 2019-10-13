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
)

// GatekeeperEnvironment : Codewind Gatekeeper Environment
type GatekeeperEnvironment struct {
	AuthURL  string `json:"auth_url"`
	Realm    string `json:"realm"`
	ClientID string `json:"client_id"`
}

// GetGatekeeperEnvironment : Fetch the Gatekeeper environment
func GetGatekeeperEnvironment(host string) (*GatekeeperEnvironment, error) {

	resp, err := http.Get(host + "/api/v1/gatekeeper/environment")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var environment GatekeeperEnvironment
	err = json.Unmarshal(byteArray, &environment)
	if err != nil {
		return nil, err
	}
	return &environment, nil
}
