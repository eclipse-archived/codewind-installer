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

package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testProjectID = "a9384430-f177-11e9-b862-edc28aca827a"
const testConnectionID = "local"

// Test_ProjectConnection :  Tests
func Test_ProjectConnection(t *testing.T) {
	ResetConnectionFile(testProjectID)

	t.Run("Asserts there are no target connections", func(t *testing.T) {
		connection, projError := GetConnection(testProjectID)
		if projError != nil {
			t.Fail()
		}
		assert.Equal(t, connection.ID, "")
	})

	t.Run("Asserts getting connection URL fails", func(t *testing.T) {
		_, projError := GetConnectionURL(testProjectID)
		if projError == nil {
			t.Fail()
		}
		assert.Equal(t, errOpConNotFound, projError.Op)
	})

	t.Run("Add project to local connection", func(t *testing.T) {
		projError := SetConnection(testProjectID, testConnectionID)
		if projError != nil {
			t.Fail()
		}
	})

	t.Run("Asserts re-adding the same connection succeeds", func(t *testing.T) {
		projError := SetConnection(testProjectID, testConnectionID)
		if projError != nil {
			t.Fail()
		}
	})

	t.Run("Asserts there is just 1 target connection added", func(t *testing.T) {
		connection, projError := GetConnection(testProjectID)
		if projError != nil {
			t.Fail()
		}
		assert.Equal(t, connection.ID, testConnectionID)
	})

	t.Run("Asserts removing a known connection is successful", func(t *testing.T) {
		projError := ResetConnectionFile(testProjectID)
		if projError != nil {
			t.Fail()
		}
	})

	t.Run("Asserts there are no targets left for this project", func(t *testing.T) {
		connection, projError := GetConnection(testProjectID)
		if projError != nil {
			t.Fail()
		}
		assert.Equal(t, connection.ID, "")
	})

	t.Run("Asserts attempting to manage an invalid project ID fails", func(t *testing.T) {
		projError := SetConnection("bad-project-ID", testConnectionID)
		if projError == nil {
			t.Fail()
		}
		assert.Equal(t, errOpInvalidID, projError.Op)
	})

}
