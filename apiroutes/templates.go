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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/eclipse/codewind-installer/config"
	"github.com/eclipse/codewind-installer/utils"
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

	// RepoOperation represents a requested operation on a template repository.
	RepoOperation struct {
		Operation string `json:"op"`
		URL       string `json:"url"`
		Value     string `json:"value"`
	}

	// SubResponseFromBatchOperation represents a sub-response
	// to a requested operation on a template repository.
	SubResponseFromBatchOperation struct {
		Status             int           `json:"status"`
		RequestedOperation RepoOperation `json:"requestedOperation"`
		Error              string        `json:"error"`
	}
)

// GetTemplates gets project templates from PFE's REST API.
// Filter them using the function arguments
func GetTemplates(projectStyle string, showEnabledOnly string) ([]Template, error) {
	req, err := http.NewRequest("GET", config.PFEApiRoute()+"templates", nil)
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
	resp, err := http.Get(config.PFEApiRoute() + "templates/styles")
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
func GetTemplateRepos() ([]utils.TemplateRepo, error) {
	resp, err := http.Get(config.PFEApiRoute() + "templates/repositories")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repos []utils.TemplateRepo
	json.Unmarshal(byteArray, &repos)

	return repos, nil
}

// AddTemplateRepo adds a template repo to PFE and
// returns the new list of existing repos
func AddTemplateRepo(URL, description string, name string) ([]utils.TemplateRepo, error) {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
	}

	values := map[string]string{
		"url":         URL,
		"description": description,
		"name":        name,
	}
	jsonValue, _ := json.Marshal(values)

	resp, err := http.Post(
		config.PFEApiRoute()+"templates/repositories",
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

	var repos []utils.TemplateRepo
	json.Unmarshal(byteArray, &repos)

	return repos, nil
}

// DeleteTemplateRepo deletes a template repo from PFE and
// returns the new list of existing repos
func DeleteTemplateRepo(URL string) ([]utils.TemplateRepo, error) {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
	}

	values := map[string]string{"url": URL}
	jsonValue, _ := json.Marshal(values)

	req, err := http.NewRequest(
		"DELETE",
		config.PFEApiRoute()+"templates/repositories",
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

	var repos []utils.TemplateRepo
	json.Unmarshal(byteArray, &repos)

	return repos, nil
}

// EnableTemplateRepos enables a template repo in PFE and
// returns the new list of template repos
func EnableTemplateRepos(repoURLs []string) ([]utils.TemplateRepo, error) {
	if repoURLs == nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", repoURLs)
	}

	var operations []RepoOperation
	for _, URL := range repoURLs {
		if _, err := url.ParseRequestURI(URL); err != nil {
			return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
		}
		operation := RepoOperation{
			Operation: "enable",
			URL:       URL,
			Value:     "true",
		}
		operations = append(operations, operation)
	}
	_, err := BatchPatchTemplateRepos(operations)
	if err != nil {
		return nil, err
	}

	repos, err := GetTemplateRepos()
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// DisableTemplateRepos enables a template repo in PFE and
// returns the new list of template repos
func DisableTemplateRepos(repoURLs []string) ([]utils.TemplateRepo, error) {
	if repoURLs == nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", repoURLs)
	}

	var operations []RepoOperation
	for _, URL := range repoURLs {
		if _, err := url.ParseRequestURI(URL); err != nil {
			return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
		}
		operation := RepoOperation{
			Operation: "enable",
			URL:       URL,
			Value:     "false",
		}
		operations = append(operations, operation)
	}
	_, err := BatchPatchTemplateRepos(operations)
	if err != nil {
		return nil, err
	}

	repos, err := GetTemplateRepos()
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// BatchPatchTemplateRepos requests that PFE perform batch operations on template repositories and
// returns a list of sub-responses to the requested operations
func BatchPatchTemplateRepos(operations []RepoOperation) ([]SubResponseFromBatchOperation, error) {
	jsonValue, _ := json.Marshal(operations)

	req, err := http.NewRequest(
		"PATCH",
		config.PFEApiRoute()+"batch/templates/repositories",
		bytes.NewBuffer(jsonValue),
	)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 207 {
		return nil, fmt.Errorf("Error: PFE responded with status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var subResponsesFromBatchOperation []SubResponseFromBatchOperation
	json.Unmarshal(byteArray, &subResponsesFromBatchOperation)

	return subResponsesFromBatchOperation, nil
}
