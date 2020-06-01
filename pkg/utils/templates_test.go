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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractGitCredentials_Fail(t *testing.T) {
	tests := map[string]struct {
		inUsername            string
		inPassword            string
		inPersonalAccessToken string
		wantErrMsg            string
	}{
		"multiple auth methods": {
			inUsername:            "username",
			inPassword:            "",
			inPersonalAccessToken: "personalAccessToken",
			wantErrMsg:            "received credentials for multiple authentication methods",
		},
		"username but no password": {
			inUsername:            "username",
			inPassword:            "",
			inPersonalAccessToken: "",
			wantErrMsg:            "received username but no password",
		},
		"password but no username": {
			inUsername:            "",
			inPassword:            "password",
			inPersonalAccessToken: "",
			wantErrMsg:            "received password but no username",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ExtractGitCredentials(test.inUsername, test.inPassword, test.inPersonalAccessToken)
			assert.Nil(t, got)
			assert.Equal(t, err.Error(), test.wantErrMsg)
		})
	}
}
