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

package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Template represents a template.
type Template struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Language    string `json:"language"`
	URL         string `json:"url"`
	ProjectType string `json:"projectType"`
}

// ListTemplates lists all templates Codewind is aware of.
func ListTemplates() {
	templates, err := GetTemplates()
	if err != nil {
		fmt.Printf("Error getting templates: %q", err)
		return
	}
	PrettyPrintJSON(templates)
}

// GetTemplates extracts URLs from the file at the provided path,
// then gets the template descriptions from those URLs,
// then formats the descriptions into template objects.
func GetTemplates() ([]Template, error) {
	resp, err := http.Get("http://localhost:9090/api/v1/templates")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var templates []Template
	json.Unmarshal(byteArray, &templates)

	return templates, nil
}

// PrettyPrintJSON prints JSON prettily.
func PrettyPrintJSON(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}
