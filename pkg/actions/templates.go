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
	"strings"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

// ListTemplates lists project templates of which Codewind is aware.
// Filter them by providing flags
func ListTemplates(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	templates, err := apiroutes.GetTemplates(
		c.String("projectStyle"),
		c.Bool("showEnabledOnly"),
		conID,
	)
	if err != nil {
		log.Printf("Error getting templates: %q", err)
		return
	}
	if len(templates) > 0 {
		PrettyPrintJSON(templates)
	} else {
		fmt.Println(templates)
	}

}

// ListTemplateStyles lists all template styles of which Codewind is aware.
func ListTemplateStyles(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	styles, err := apiroutes.GetTemplateStyles(conID)
	if err != nil {
		log.Printf("Error getting template styles: %q", err)
		return
	}
	PrettyPrintJSON(styles)
}

// ListTemplateRepos lists all template repos of which Codewind is aware.
func ListTemplateRepos(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, err := apiroutes.GetTemplateRepos(conID)
	if err != nil {
		log.Printf("Error getting template repos: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// AddTemplateRepo adds the provided template repo to PFE.
func AddTemplateRepo(c *cli.Context) {
	url := c.String("url")
	repos, err := apiroutes.AddTemplateRepo(
		url,
		c.String("description"),
		c.String("name"),
		strings.TrimSpace(strings.ToLower(c.String("conid"))),
	)
	if err != nil {
		log.Printf("Error adding template repo: %q", err)
		return
	}
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	extensions, err := apiroutes.GetExtensions(conID)
	if err == nil {
		utils.OnAddTemplateRepo(extensions, url, repos)
	}
	PrettyPrintJSON(repos)
}

// DeleteTemplateRepo deletes the provided template repo from PFE.
func DeleteTemplateRepo(c *cli.Context) {
	url := c.String("url")
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	extensions, err := apiroutes.GetExtensions(conID)
	if err == nil {
		repos, err2 := apiroutes.GetTemplateRepos(conID)
		if err2 == nil {
			utils.OnDeleteTemplateRepo(extensions, url, repos)
		}
	}
	repos, err := apiroutes.DeleteTemplateRepo(url, conID)
	if err != nil {
		log.Printf("Error deleting template repo: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// EnableTemplateRepos enables templates repo of which Codewind is aware.
func EnableTemplateRepos(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, err := apiroutes.EnableTemplateRepos(c.Args(), conID)
	if err != nil {
		log.Printf("Error enabling template repos: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// DisableTemplateRepos disables templates repo of which Codewind is aware.
func DisableTemplateRepos(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, err := apiroutes.DisableTemplateRepos(c.Args(), conID)
	if err != nil {
		log.Printf("Error enabling template repos: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// PrettyPrintJSON prints JSON prettily.
func PrettyPrintJSON(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}
