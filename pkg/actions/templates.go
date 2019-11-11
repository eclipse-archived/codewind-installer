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
	"strings"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// ListTemplates lists project templates of which Codewind is aware.
// Filter them by providing flags
func ListTemplates(c *cli.Context) {
	projectStyle := c.String("projectStyle")
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	showEnabledOnly := c.Bool("showEnabledOnly")
	templates, templatesErr := apiroutes.GetTemplates(conID, projectStyle, showEnabledOnly)
	if templatesErr != nil {
		logr.Errorf("Error getting templates: %q", templatesErr)
		return
	}
	if len(templates) > 0 {
		PrettyPrintJSON(templates)
	} else {
		logr.Infoln(templates)
	}

}

// ListTemplateStyles lists all template styles of which Codewind is aware.
func ListTemplateStyles(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	styles, err := apiroutes.GetTemplateStyles(conID)
	if err != nil {
		logr.Errorf("Error getting template styles: %q", err)
		return
	}
	PrettyPrintJSON(styles)
}

// ListTemplateRepos lists all template repos of which Codewind is aware.
func ListTemplateRepos(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, err := apiroutes.GetTemplateRepos(conID)
	if err != nil {
		logr.Errorf("Error getting template repos: %q", err)
		return
	}
	PrettyPrintJSON(repos)
}

// AddTemplateRepo adds the provided template repo to PFE.
func AddTemplateRepo(c *cli.Context) {
	url := c.String("url")
	desc := c.String("description")
	name := c.String("name")
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, err := apiroutes.AddTemplateRepo(conID, url, desc, name)
	if err != nil {
		logr.Errorf("Error adding template repo: %q", err)
		return
	}
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
	extensions, extensionsErr := apiroutes.GetExtensions(conID)
	if extensionsErr == nil {
		repos, reposErr := apiroutes.GetTemplateRepos(conID)
		if reposErr == nil {
			utils.OnDeleteTemplateRepo(extensions, url, repos)
		}
	}
	repos, reposErr := apiroutes.DeleteTemplateRepo(conID, url)
	if reposErr != nil {
		logr.Errorf("Error deleting template repo: %q", reposErr)
		return
	}
	PrettyPrintJSON(repos)
}

// EnableTemplateRepos enables templates repo of which Codewind is aware.
func EnableTemplateRepos(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, reposErr := apiroutes.EnableTemplateRepos(conID, c.Args())
	if reposErr != nil {
		logr.Errorf("Error enabling template repos: %q", reposErr)
		return
	}
	PrettyPrintJSON(repos)
}

// DisableTemplateRepos disables templates repo of which Codewind is aware.
func DisableTemplateRepos(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, reposErr := apiroutes.DisableTemplateRepos(conID, c.Args())
	if reposErr != nil {
		logr.Errorf("Error enabling template repos: %q", reposErr)
		return
	}
	PrettyPrintJSON(repos)
}

// PrettyPrintJSON prints JSON prettily.
func PrettyPrintJSON(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	logr.Infoln(string(s))
}
