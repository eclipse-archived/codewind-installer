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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsPFEReady(t *testing.T) {
	t.Run("Asserts PFE ready", func(t *testing.T) {
		mockClientTrue := &MockResponse{StatusCode: http.StatusOK, Body: nil}
		PFEReady, err := IsPFEReady(mockClientTrue, "http://test-connection.com")
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, PFEReady, true)
	})
	t.Run("Asserts PFE not ready", func(t *testing.T) {
		mockClientFalse := &MockResponse{StatusCode: http.StatusNotFound, Body: nil}
		PFENotReady, err := IsPFEReady(mockClientFalse, "http://test-connection.com")
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, PFENotReady, false)
	})
}
