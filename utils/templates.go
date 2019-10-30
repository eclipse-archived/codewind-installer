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

package utils

type (
	// TemplateRepo represents a template repository.
	TemplateRepo struct {
		Description   string   `json:"description"`
		URL           string   `json:"url"`
		Name          string   `json:"name"`
		ID            string   `json:"id"`
		Enabled       bool     `json:"enabled"`
		Protected     bool     `json:"protected"`
		ProjectStyles []string `json:"projectStyles"`
	}
)

func getApplicableCommand(extension Extension, repo TemplateRepo, name string) *ExtensionCommand {

	// does extension specify a style?
	style := extension.Config.Style
	if style == "" {
		return nil
	}

	// determine if repository has a matching style
	// if so, look for an applicable command
	for _, s := range repo.ProjectStyles {
		if s == style {
			for _, command := range extension.Commands {
				if command.Name == name {
					return &command
				}
			}
			break
		}
	}

	return nil
}

func onRepositoryAdd(extensions []Extension, repo TemplateRepo) {

	if repo.ID == "" {
		return
	}

	for _, extension := range extensions {
		cmdPtr := getApplicableCommand(extension, repo, "onRepositoryAdd")
		if cmdPtr != nil {
			params := make(map[string]string)
			params["$id"] = repo.ID
			params["$url"] = repo.URL
			RunCommand("", *cmdPtr, params)
		}
	}
}

func onRepositoryRemove(extensions []Extension, repo TemplateRepo) {

	if repo.ID == "" {
		return
	}

	for _, extension := range extensions {
		cmdPtr := getApplicableCommand(extension, repo, "onRepositoryRemove")
		if cmdPtr != nil {
			params := make(map[string]string)
			params["$id"] = repo.ID
			RunCommand("", *cmdPtr, params)
		}
	}
}

// OnAddTemplateRepo runs any extension command associated with a repo add
func OnAddTemplateRepo(extensions []Extension, url string, repos []TemplateRepo) {
	// look for what was just added
	for _, repo := range repos {
		if repo.URL == url {
			onRepositoryAdd(extensions, repo)
			break
		}
	}
}

// OnDeleteTemplateRepo runs any extension command associated with a repo delete
func OnDeleteTemplateRepo(extensions []Extension, url string, repos []TemplateRepo) {
	// look for what's to be deleted
	for _, repo := range repos {
		if repo.URL == url {
			onRepositoryRemove(extensions, repo)
			break
		}
	}
}
