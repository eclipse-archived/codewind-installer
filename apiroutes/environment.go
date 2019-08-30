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
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/urfave/cli"
)

type Environment struct {
	RunningOnICP      bool   `json:"running_on_icp"`
	UserString        string `json:"user_string"`
	SocketNamespace   string `json:"socket_namespace"`
	Version           string `json:"codewind_version"`
	WorkspaceLocation string `json:"workspace_location"`
	Platform          string `json:"os_platform"`
}

func GetAPIEnvironment(c *cli.Context, host string) (*Environment, error) {

	if c.GlobalBool("insecure") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	resp, err := http.Get(host + "api/v1/environment")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var environment Environment
	err = json.Unmarshal(byteArray, &environment)
	if err != nil {
		return nil, err
	}
	return &environment, nil
}
