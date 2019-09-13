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
)

type (
	// ExtensionCommand represents a command defined by a project extension
	ExtensionCommand struct {
		Name    string `json:"name"`
		Command string `json:"command"`
	}

	// Extension represents a project extension defined by codewind.yaml
	Extension struct {
		ProjectType string             `json:"projectType"`
		Detection   string             `json:"detection"`
		Commands    []ExtensionCommand `json:"commands"`
	}
)

// GetExtensions gets project extensions from PFE's REST API.
func GetExtensions() ([]Extension, error) {
	resp, err := http.Get(config.PFEApiRoute() + "extensions")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var extensions []Extension
	json.Unmarshal(byteArray, &extensions)

	return extensions, nil
}
