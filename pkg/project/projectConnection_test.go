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
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/stretchr/testify/assert"
)

const testProjectID = "a9384430-f177-11e9-b862-edc28aca827a"
const testConnectionID = "testCon"
const testHost = "http://test-host"
const schemaVersion = 1

func WriteNewConfigFile() error {
	connectionsFile := connections.ConnectionConfig{
		SchemaVersion: schemaVersion,
		Connections: []connections.Connection{
			connections.Connection{
				ID:       "local",
				Label:    "Codewind local connection",
				URL:      "",
				AuthURL:  "",
				Realm:    "",
				ClientID: "",
			},
			connections.Connection{
				ID:       "testCon",
				Label:    "Test remote connection",
				URL:      testHost,
				AuthURL:  "",
				Realm:    "",
				ClientID: "",
			},
		},
	}
	body, err := json.MarshalIndent(connectionsFile, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(connections.GetConnectionConfigFilename(), body, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Test_ProjectConnection :  Tests
func Test_ProjectConnection(t *testing.T) {
	WriteNewConfigFile()

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	t.Run("Asserts project connection file doesn't exist", func(t *testing.T) {
		connectionExists := ConnectionFileExists(testProjectID)
		assert.Equal(t, false, connectionExists)
	})

	t.Run("Asserts setting a default connection file doesn't fail", func(t *testing.T) {
		projError := CreateConnectionFile(testProjectID)
		if projError != nil {
			t.Fail()
		}
	})

	t.Run("Asserts project connection file exists", func(t *testing.T) {
		connectionExists := ConnectionFileExists(testProjectID)
		assert.Equal(t, true, connectionExists)
	})

	t.Run("Asserts project defaults to local connection", func(t *testing.T) {
		connection, projError := GetConnection(testProjectID)
		if projError != nil {
			t.Fail()
		}
		assert.Equal(t, "local", connection.ID)
	})

	t.Run("Asserts a new connectionID can be set", func(t *testing.T) {
		projError := SetConnection(testConnectionID, testProjectID)
		if projError != nil {
			t.Fail()
		}
	})

	t.Run("Asserts the correct connection has been added", func(t *testing.T) {
		connection, projError := GetConnection(testProjectID)
		if projError != nil {
			t.Fail()
		}
		assert.Equal(t, testConnectionID, connection.ID)
	})

	t.Run("Asserts resetting the connection is successful", func(t *testing.T) {
		projError := ResetConnectionFile(testProjectID)
		if projError != nil {
			t.Fail()
		}
	})

	t.Run("Asserts connection is reset to local", func(t *testing.T) {
		connection, projError := GetConnection(testProjectID)
		if projError != nil {
			t.Fail()
		}
		assert.Equal(t, "local", connection.ID)
	})

	t.Run("Asserts attempting to set an invalid project ID fails", func(t *testing.T) {
		projError := SetConnection(testConnectionID, "bad-project-ID")
		if projError == nil {
			t.Fail()
		}
		assert.Equal(t, errOpInvalidID, projError.Op)
	})

	t.Run("Asserts the connection file can be removed", func(t *testing.T) {
		projError := RemoveConnectionFile(testProjectID)
		if projError != nil {
			t.Fail()
		}
	})

	t.Run("Asserts the connection file has been removed", func(t *testing.T) {
		connectionExists := ConnectionFileExists(testProjectID)
		assert.Equal(t, false, connectionExists)
	})
	connections.ResetConnectionsFile()
}
