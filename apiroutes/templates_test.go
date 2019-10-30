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
	"errors"
	"log"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var numCodewindTemplates int = 8
var numAppsodyTemplates int = 11
var numTemplates int = numCodewindTemplates + numAppsodyTemplates

func TestGetTemplates(t *testing.T) {
	tests := map[string]struct {
		inProjectStyle    string
		inShowEnabledOnly string
		wantedType        []Template
		wantedLength      int
	}{
		"get templates of all styles": {
			inProjectStyle:    "",
			inShowEnabledOnly: "",
			wantedType:        []Template{},
			wantedLength:      numTemplates,
		},
		"filter templates by known style": {
			inProjectStyle: "Codewind",
			wantedType:     []Template{},
			wantedLength:   numCodewindTemplates,
		},
		"filter templates by unknown style": {
			inProjectStyle: "unknownStyle",
			wantedType:     []Template{},
			wantedLength:   0,
		},
		"filter templates by enabled templates": {
			inShowEnabledOnly: "true",
			wantedType:        []Template{},
			wantedLength:      numTemplates,
		},
		"filter templates by enabled templates of unknown style": {
			inProjectStyle:    "unknownStyle",
			inShowEnabledOnly: "false",
			wantedType:        []Template{},
			wantedLength:      0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplates(test.inProjectStyle, test.inShowEnabledOnly)
			assert.IsType(t, test.wantedType, got)
			assert.Equal(t, test.wantedLength, len(got))
			assert.Nil(t, err)
		})
	}
}

func TestGetTemplateStyles(t *testing.T) {
	tests := map[string]struct {
		want      []string
		wantedErr error
	}{
		"success case": {
			want:      []string{"Appsody", "Codewind"},
			wantedErr: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplateStyles()
			assert.Equal(t, test.want, got)
			assert.IsType(t, test.wantedErr, err)
		})
	}
}

func TestGetTemplateRepos(t *testing.T) {
	tests := map[string]struct {
		wantedType   []TemplateRepo
		wantedLength int
		wantedErr    error
	}{
		"success case": {
			wantedType:   []TemplateRepo{},
			wantedLength: 3,
			wantedErr:    nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplateRepos()
			assert.IsType(t, test.wantedType, got)
			assert.Equal(t, test.wantedLength, len(got))
			assert.Equal(t, test.wantedErr, err)
		})
	}
}

func TestFailuresAddTemplateRepo(t *testing.T) {
	tests := map[string]struct {
		inURL         string
		inDescription string
		wantedType    []TemplateRepo
		wantedErr     error
	}{
		"fail case: add invalid URL": {
			inURL:         "invalidURL",
			inDescription: "invalidURL",
			wantedType:    nil,
			wantedErr:     errors.New("Error: 'invalidURL' is not a valid URL"),
		},
		"fail case: add duplicate URL": {
			inURL:         "https://raw.githubusercontent.com/kabanero-io/codewind-templates/master/devfiles/index.json",
			inDescription: "example repository containing links to templates",
			wantedType:    nil,
			wantedErr:     errors.New("Error: PFE responded with status code 400"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := AddTemplateRepo(test.inURL, test.inDescription, "template-name")
			assert.IsType(t, test.wantedType, got, "got: %v", got)
			assert.Equal(t, test.wantedErr, err)
		})
	}
}

func TestFailuresDeleteTemplateRepo(t *testing.T) {
	tests := map[string]struct {
		inURL      string
		wantedType []TemplateRepo
		wantedErr  error
	}{
		"fail case: remove invalid URL": {
			inURL:      "invalidURL",
			wantedType: nil,
			wantedErr:  new(url.Error),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := DeleteTemplateRepo(test.inURL)
			assert.IsType(t, test.wantedType, got, "got: %v", got)
			assert.IsType(t, test.wantedErr, err)
		})
	}
}

func TestSuccessfulAddAndDeleteTemplateRepo(t *testing.T) {
	testRepoURL := "https://raw.githubusercontent.com/kabanero-io/codewind-templates/aad4bafc14e1a295fb8e462c20fe8627248609a3/devfiles/index.json"

	originalRepos, err := GetTemplateRepos()
	if err != nil {
		log.Fatalf("[TestSuccessfulAddAndDeleteTemplateRepo] Error getting template repos: %s", err)
	}
	originalNumRepos := len(originalRepos)

	t.Run("Successfully add template repo", func(t *testing.T) {
		wantedNumRepos := originalNumRepos + 1

		got, err := AddTemplateRepo(testRepoURL, "example description", "template-name")

		assert.IsType(t, []TemplateRepo{}, got)
		assert.Equal(t, wantedNumRepos, len(got), "got: %v", got)
		assert.Nil(t, err)
	})

	t.Run("Successfully delete template repo", func(t *testing.T) {
		wantedNumRepos := originalNumRepos

		got, err := DeleteTemplateRepo(testRepoURL)

		assert.IsType(t, []TemplateRepo{}, got)
		assert.Equal(t, wantedNumRepos, len(got), "got: %v", got)
		assert.Nil(t, err)
	})

	// This test cleans up after itself
}
