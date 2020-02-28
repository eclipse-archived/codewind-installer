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
	"fmt"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

// ListTemplates lists project templates of which Codewind is aware.
// Filter them by providing flags
func ListTemplates(c *cli.Context) {
	projectStyle := c.String("projectStyle")
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	showEnabledOnly := c.Bool("showEnabledOnly")
	templates, err := apiroutes.GetTemplates(conID, projectStyle, showEnabledOnly)
	if err != nil {
		templateErr := &TemplateError{errOpListTemplates, err, err.Error()}
		HandleTemplateError(templateErr)
		return
	}
	if len(templates) > 0 {
		utils.PrettyPrintJSON(templates)
	} else {
		fmt.Println(templates)
	}
}

// ListTemplateStyles lists all template styles of which Codewind is aware.
func ListTemplateStyles(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	styles, err := apiroutes.GetTemplateStyles(conID)
	if err != nil {
		templateErr := &TemplateError{errOpListStyles, err, err.Error()}
		HandleTemplateError(templateErr)
		return
	}
	utils.PrettyPrintJSON(styles)
}

// ListTemplateRepos lists all template repos of which Codewind is aware.
func ListTemplateRepos(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, err := apiroutes.GetTemplateRepos(conID)
	if err != nil {
		templateErr := &TemplateError{errOpListRepos, err, err.Error()}
		HandleTemplateError(templateErr)
		return
	}
	utils.PrettyPrintJSON(repos)
}

// AddTemplateRepo adds the provided template repo to PFE.
func AddTemplateRepo(c *cli.Context) {
	url := c.String("url")
	desc := c.String("description")
	name := c.String("name")
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, err := apiroutes.AddTemplateRepo(conID, url, desc, name)
	if err != nil {
		templateErr := &TemplateError{errOpAddRepo, err, err.Error()}
		HandleTemplateError(templateErr)
		return
	}
	extensions, err := apiroutes.GetExtensions(conID)
	if err == nil {
		utils.OnAddTemplateRepo(extensions, url, repos)
	}
	utils.PrettyPrintJSON(repos)
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
	repos, err := apiroutes.DeleteTemplateRepo(conID, url)
	if err != nil {
		templateErr := &TemplateError{errOpDeleteRepo, err, err.Error()}
		HandleTemplateError(templateErr)
		return
	}
	utils.PrettyPrintJSON(repos)
}

// EnableTemplateRepos enables templates repo of which Codewind is aware.
func EnableTemplateRepos(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, err := apiroutes.EnableTemplateRepos(conID, c.Args())
	if err != nil {
		templateErr := &TemplateError{errOpEnableRepo, err, err.Error()}
		HandleTemplateError(templateErr)
		return
	}
	utils.PrettyPrintJSON(repos)
}

// DisableTemplateRepos disables templates repo of which Codewind is aware.
func DisableTemplateRepos(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	repos, err := apiroutes.DisableTemplateRepos(conID, c.Args())
	if err != nil {
		templateErr := &TemplateError{errOpDisableRepo, err, err.Error()}
		HandleTemplateError(templateErr)
		return
	}
	utils.PrettyPrintJSON(repos)
}
