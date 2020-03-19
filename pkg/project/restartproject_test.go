/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

package project

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/stretchr/testify/assert"
)

var emptyResponseBody = ioutil.NopCloser(bytes.NewReader([]byte{}))
var mockConnection = connections.Connection{ID: "local"}

func Test_RestartProject(t *testing.T) {
	t.Run("success case - returns nil error when PFE status code 202", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusAccepted, Body: emptyResponseBody}
		err := RestartProject(mockClient, &mockConnection, "mockURL", "mockID", "debugNoInit")
		assert.Nil(t, err)
	})

	t.Run("error case - returns error when PFE status code non 202", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusNotFound, Body: emptyResponseBody}
		err := RestartProject(mockClient, &mockConnection, "mockURL", "mockID", "debugNoInit")
		assert.NotNil(t, err)
	})
}
