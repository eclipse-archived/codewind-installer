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

	"github.com/eclipse/codewind-installer/config"
)

// Template represents a project template.
type Template struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Language    string `json:"language"`
	URL         string `json:"url"`
	ProjectType string `json:"projectType"`
}

// ListTemplates lists all project templates Codewind of which is aware.
func ListTemplates() {
	templates, err := GetTemplates()
	if err != nil {
		fmt.Printf("Error getting templates: %q", err)
		return
	}
	PrettyPrintJSON(templates)
}


// ListTemplateStyles lists all template styles Codewind of which is aware.
func ListTemplateStyles() {
	styles, err := GetTemplateStyles()
	if err != nil {
		fmt.Printf("Error getting template styles: %q", err)
		return
	}
	PrettyPrintJSON(styles)
}

// GetTemplates gets all project templates from PFE's REST API
func GetTemplates() ([]Template, error) {
	resp, err := http.Get(config.PFEApiRoute + "templates")
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

// GetTemplateStyles gets all template styles from PFE's REST API
func GetTemplateStyles() ([]string, error) {
	resp, err := http.Get(config.PFEApiRoute + "templates/styles")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var styles []string
	json.Unmarshal(byteArray, &styles)

	return styles, nil
}

// PrettyPrintJSON prints JSON prettily.
func PrettyPrintJSON(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}
