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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplates(t *testing.T) {
	tests := map[string]struct {
		inProjectStyle string
		inShowEnabledOnly string
		wantedType     []Template
		wantedLength   int
	}{
		"get templates of all styles": {
			inProjectStyle: "",
			inShowEnabledOnly: "",
			wantedType:     []Template{},
			wantedLength:   8,
		},
		"filter templates by known style": {
			inProjectStyle: "Codewind",
			wantedType:   []Template{},
			wantedLength: 8,
		},
		"filter templates by unknown style": {
			inProjectStyle: "Appsody",
			wantedType:   []Template{},
			wantedLength: 0,
		},
		"filter templates by enabled templates": {
			inShowEnabledOnly: "true",
			wantedType:   []Template{},
			wantedLength: 8,
		},
		"filter templates by enabled templates of unknown style": {
			inProjectStyle: "Appsody",
			inShowEnabledOnly: "false",
			wantedType:     []Template{},
			wantedLength:   0,
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
			want:      []string{"Codewind"},
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
			wantedLength: 1,
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
