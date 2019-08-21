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
	"log"
	"net/http"

	"github.com/eclipse/codewind-installer/config"
)

type (
	// Template represents a project template.
	Template struct {
		Label       string `json:"label"`
		Description string `json:"description"`
		Language    string `json:"language"`
		URL         string `json:"url"`
		ProjectType string `json:"projectType"`
	}

	// TemplateRepo represents a template repository.
	TemplateRepo struct {
		Description string `json:"description"`
		URL         string `json:"url"`
	}
)

// ListTemplates lists all project templates of which Codewind is aware.
func ListTemplates() {
	templates, err := GetTemplates()
	if err != nil {
		log.Printf("Error getting templates: %q", err)
		return
	}
	PrettyPrintJSON(templates)
}

// ListTemplateStyles lists all template styles of which Codewind is aware.
func ListTemplateStyles() {
	styles, err := GetTemplateStyles()
	if err != nil {
		log.Printf("Error getting template styles: %q", err)
		return
	}
	PrettyPrintJSON(styles)
}

// ListTemplateRepos lists all template repos of which Codewind is aware.
func ListTemplateRepos() {
	repos, err := GetTemplateRepos()
	if err != nil {
		fmt.Printf("Error getting template repos: %q", err)
		return
	}
	PrettyPrintJSON(repos)
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

// GetTemplateRepos gets all template repos from PFE's REST API
func GetTemplateRepos() ([]TemplateRepo, error) {
	resp, err := http.Get(config.PFEApiRoute + "templates/repositories")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repos []TemplateRepo
	json.Unmarshal(byteArray, &repos)

	return repos, nil
}

// PrettyPrintJSON prints JSON prettily.
func PrettyPrintJSON(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}
