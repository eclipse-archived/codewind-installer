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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/eclipse/codewind-installer/config"
	"github.com/urfave/cli"
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

// ListTemplates lists project templates of which Codewind is aware.
// Filter them by providing flags
func ListTemplates(c *cli.Context) {
	templates, err := GetTemplates(
		c.String("projectStyle"),
		c.String("showEnabledOnly"),
	)
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
		log.Printf("Error getting template repos: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// CmdToAddTemplateRepo adds the provided template repo to PFE.
func CmdToAddTemplateRepo(c *cli.Context) {
	repos, err := AddTemplateRepo(
		c.String("URL"),
		c.String("description"),
	)
	if err != nil {
		log.Printf("Error adding template repo: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// CmdToDeleteTemplateRepo deletes the provided template repo from PFE.
func CmdToDeleteTemplateRepo(c *cli.Context) {
	repos, err := DeleteTemplateRepo(c.String("URL"))
	if err != nil {
		log.Printf("Error deleting template repo: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// GetTemplates gets project templates from PFE's REST API.
// Filter them using the function arguments
func GetTemplates(projectStyle string, showEnabledOnly string) ([]Template, error) {
	req, err := http.NewRequest("GET", config.PFEApiRoute+"templates", nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	if projectStyle != "" {
		query.Add("projectStyle", projectStyle)
	}
	if showEnabledOnly != "" {
		query.Add("showEnabledOnly", showEnabledOnly)
	}
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
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

// AddTemplateRepo adds a template repo to PFE and
// returns the new list of existing repos
func AddTemplateRepo(URL, description string) ([]TemplateRepo, error) {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
	}

	values := map[string]string{
		"url":         URL,
		"description": description,
	}
	jsonValue, _ := json.Marshal(values)

	resp, err := http.Post(
		config.PFEApiRoute+"templates/repositories",
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: PFE responded with status code %d", resp.StatusCode)
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

// DeleteTemplateRepo deletes a template repo from PFE and
// returns the new list of existing repos
func DeleteTemplateRepo(URL string) ([]TemplateRepo, error) {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return nil, err
	}

	values := map[string]string{"url": URL}
	jsonValue, _ := json.Marshal(values)

	req, err := http.NewRequest(
		"DELETE",
		config.PFEApiRoute+"templates/repositories",
		bytes.NewBuffer(jsonValue),
	)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: PFE responded with status code %d", resp.StatusCode)
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
