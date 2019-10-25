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

package deployments

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/eclipse/codewind-installer/apiroutes"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

type ClientMockServerConfig struct {
	StatusCode int
	Body       io.ReadCloser
}

func (c *ClientMockServerConfig) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: c.StatusCode,
		Body:       c.Body,
	}, nil
}

// Test_SchemaUpgrade01 :  Upgrade schema tests from Version 0 to Version 1
func Test_SchemaUpgrade0to1(t *testing.T) {
	// create a v1 file :
	v1File := "{\"active\": \"testlocal\",\"deployments\": [{\"name\":\"testlocal\",\"label\": \"Codewind local test deployment\",\"url\": \"\"}]}"
	ioutil.WriteFile(getDeploymentConfigFilename(), []byte(v1File), 0644)
	t.Run("Asserts schema updated to v1 with a local target", func(t *testing.T) {
		InitConfigFileIfRequired() // perform upgrade
		result, err := GetDeploymentsConfig()
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, 1, result.SchemaVersion)
		assert.Equal(t, "testlocal", result.Active)
		assert.Len(t, result.Deployments, 1)
		assert.Equal(t, "testlocal", result.Deployments[0].ID)
	})
}

func Test_GetDeploymentsConfig(t *testing.T) {
	t.Run("Asserts there is only one deployment", func(t *testing.T) {
		ResetDeploymentsFile()
		result, err := GetDeploymentsConfig()
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, "local", result.Active)
		assert.Len(t, result.Deployments, 1)
	})
}

func Test_GetActiveDeployment(t *testing.T) {
	t.Run("Asserts the initial deployment is local", func(t *testing.T) {
		ResetDeploymentsFile()
		result, err := FindTargetDeployment()
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, "local", result.ID)
		assert.Equal(t, "Codewind local deployment", result.Label)
		assert.Equal(t, "", result.URL)
	})
}

// Test_CreateNewDeployment :  Adds a new deployment to the list called remoteserver
func Test_CreateNewDeployment(t *testing.T) {
	set := flag.NewFlagSet("tests", 0)
	set.String("label", "MyRemoteServer", "just a label")
	set.String("url", "https://codewind.server.remote", "Codewind URL")
	c := cli.NewContext(nil, set, nil)

	ResetDeploymentsFile()

	mockResponse := apiroutes.GatekeeperEnvironment{AuthURL: "http://a.mock.auth.server.remote:1234", Realm: "remoteRealm", ClientID: "remoteClient"}
	jsonResponse, _ := json.Marshal(mockResponse)
	body := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

	// construct a http client with our mock canned response
	mockClient := &ClientMockServerConfig{StatusCode: http.StatusOK, Body: body}

	t.Run("Adds new deployment to the config", func(t *testing.T) {
		AddDeploymentToList(mockClient, c)
		result, err := GetDeploymentsConfig()
		if err != nil {
			t.Fail()
		}
		assert.Len(t, result.Deployments, 2)
	})
}

// Test_SwitchTarget : Switches the target to the last one added
func Test_SwitchTarget(t *testing.T) {

	allDeployments, err := GetAllDeployments()
	if err != nil {
		t.Fail()
	}

	newID := allDeployments[1].ID

	set := flag.NewFlagSet("tests", 0)
	set.String("depid", newID, "doc")
	c := cli.NewContext(nil, set, nil)
	t.Run("Assert target switches to remoteserver", func(t *testing.T) {
		SetTargetDeployment(c)
		result, err := FindTargetDeployment()
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, "MyRemoteServer", result.Label)
		assert.Equal(t, "https://codewind.server.remote", result.URL)
		assert.Equal(t, "http://a.mock.auth.server.remote:1234", result.AuthURL)
		assert.Equal(t, "remoteRealm", result.Realm)
		assert.Equal(t, "remoteClient", result.ClientID)
	})
}

// Test_RemoveDeploymentFromList : Adds a new deployment to the stored list
func Test_RemoveDeploymentFromList(t *testing.T) {
	set := flag.NewFlagSet("tests", 0)

	allDeployments, err := GetAllDeployments()
	if err != nil {
		t.Fail()
	}

	idToDelete := allDeployments[1].ID

	set.String("depid", idToDelete, "doc")
	c := cli.NewContext(nil, set, nil)

	t.Run("Check we have 2 deployments", func(t *testing.T) {
		result, err := GetDeploymentsConfig()
		if err != nil {
			t.Fail()
		}
		assert.Len(t, result.Deployments, 2)
	})

	t.Run("Check current target host url is https://codewind.server.remote", func(t *testing.T) {
		result, err := FindTargetDeployment()
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, "https://codewind.server.remote", result.URL)
	})

	t.Run("Remove the https://codewind.server.remote deployment", func(t *testing.T) {
		RemoveDeploymentFromList(c)
		result, err := GetDeploymentsConfig()
		if err != nil {
			t.Fail()
		}
		assert.Len(t, result.Deployments, 1)
	})

	t.Run("Check target reverts back to local", func(t *testing.T) {
		result, err := FindTargetDeployment()
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, "local", result.ID)
	})
}
