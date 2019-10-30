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
	"log"

	"github.com/eclipse/codewind-installer/apiroutes"
	"github.com/eclipse/codewind-installer/utils"
	"github.com/urfave/cli"
)

// ListTemplates lists project templates of which Codewind is aware.
// Filter them by providing flags
func ListTemplates(c *cli.Context) {
	templates, err := apiroutes.GetTemplates(
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
	styles, err := apiroutes.GetTemplateStyles()
	if err != nil {
		log.Printf("Error getting template styles: %q", err)
		return
	}
	PrettyPrintJSON(styles)
}

// ListTemplateRepos lists all template repos of which Codewind is aware.
func ListTemplateRepos() {
	repos, err := apiroutes.GetTemplateRepos()
	if err != nil {
		log.Printf("Error getting template repos: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// AddTemplateRepo adds the provided template repo to PFE.
func AddTemplateRepo(c *cli.Context) {
	url := c.String("URL")
	repos, err := apiroutes.AddTemplateRepo(
		url,
		c.String("description"),
		c.String("name"),
	)
	if err != nil {
		log.Printf("Error adding template repo: %q", err)
		return
	}
	extensions, err := apiroutes.GetExtensions()
	if err == nil {
		utils.OnAddTemplateRepo(extensions, url, repos)
	}
	PrettyPrintJSON(repos)
}

// DeleteTemplateRepo deletes the provided template repo from PFE.
func DeleteTemplateRepo(c *cli.Context) {
	url := c.String("URL")
	extensions, err := apiroutes.GetExtensions()
	if err == nil {
		repos, err2 := apiroutes.GetTemplateRepos()
		if err2 == nil {
			utils.OnDeleteTemplateRepo(extensions, url, repos)
		}
	}
	repos, err := apiroutes.DeleteTemplateRepo(url)
	if err != nil {
		log.Printf("Error deleting template repo: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// PrettyPrintJSON prints JSON prettily.
func PrettyPrintJSON(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}
