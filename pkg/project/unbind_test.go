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

func TestUnbind(t *testing.T) {
	mockConnection := connections.Connection{ID: "local"}

	body := ioutil.NopCloser(bytes.NewReader([]byte("")))
	t.Run("Expect success - project unbinds", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusAccepted, Body: body}
		err := Unbind(mockClient, &mockConnection, "dummyurl", "mockID")
		if err != nil {
			t.Errorf("Unbind() failed with error %s", err)
		}
	})

	t.Run("Expect failure - pfe returns non 202 status", func(t *testing.T) {
		mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusBadRequest, Body: body}
		err := Unbind(mockClient, &mockConnection, "dummyurl", "mockID")
		assert.Error(t, err)
	})
}
