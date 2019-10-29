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

func addRepo(repo TemplateRepo) {

	if repo.ID == "" {
		return
	}
}

// OnRepositoryAdd runs any extension command associated with a repo add
func OnRepositoryAdd(url string, extensions []Extension, repos []TemplateRepo) {
	for _, repo := range repos {
		if repo.URL == url {
			addRepo(repo)
			break
		}
	}
}

// OnRepositoryDelete runs any extension command associated with a repo delete
func OnRepositoryDelete(url string, extensions []Extension) {

}
