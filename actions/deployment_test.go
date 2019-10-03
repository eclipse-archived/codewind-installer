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
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func Test_GetDeploymentsConfig(t *testing.T) {
	InitDeploymentConfigIfRequired()
	t.Run("Asserts there is only one deployment", func(t *testing.T) {
		ResetDeploymentsFile()
		result := GetDeploymentsConfig()
		assert.Equal(t, "local", result.Active)
		assert.Len(t, result.Deployments, 1)
	})
}

func Test_GetActiveDeployment(t *testing.T) {
	t.Run("Asserts the initial deployment is local", func(t *testing.T) {
		ResetDeploymentsFile()
		result := FindTargetDeployment()
		assert.Equal(t, "local", result.ID)
		assert.Equal(t, "Codewind local deployment", result.Label)
		assert.Equal(t, "", result.URL)
	})
}

// Test_CreateNewDeployment :  Adds a new deployment to the list called remoteserver
func Test_CreateNewDeployment(t *testing.T) {
	set := flag.NewFlagSet("tests", 0)
	set.String("id", "remoteserver", "just an id")
	set.String("label", "MyRemoteServer", "just a label")
	set.String("url", "https://codewind.server.remote", "Codewind URL")
	set.String("auth", "https://auth.server.remote:8443", "Auth URL")
	set.String("realm", "codewind", "Security realm")
	set.String("clientid", "cwctl", "ID of client")

	c := cli.NewContext(nil, set, nil)
	ResetDeploymentsFile()
	t.Run("Adds new deployment to the config", func(t *testing.T) {
		AddDeploymentToList(c)
		result := GetDeploymentsConfig()
		assert.Len(t, result.Deployments, 2)
	})
}

// Test_SwitchTarget : Switches the target from one deployment to one called remoteserver
func Test_SwitchTarget(t *testing.T) {
	set := flag.NewFlagSet("tests", 0)
	set.String("id", "remoteserver", "doc")
	c := cli.NewContext(nil, set, nil)
	t.Run("Assert target switches to remoteserver", func(t *testing.T) {
		SetTargetDeployment(c)
		result := FindTargetDeployment()
		assert.Equal(t, "remoteserver", result.ID)
		assert.Equal(t, "MyRemoteServer", result.Label)
		assert.Equal(t, "https://codewind.server.remote", result.URL)
		assert.Equal(t, "https://auth.server.remote:8443", result.AuthURL)
		assert.Equal(t, "codewind", result.Realm)
		assert.Equal(t, "cwctl", result.ClientID)
	})
}

// Test_RemoveDeploymentFromList : Adds a new deployment to the stored list
func Test_RemoveDeploymentFromList(t *testing.T) {
	set := flag.NewFlagSet("tests", 0)
	set.String("id", "remoteserver", "doc")
	c := cli.NewContext(nil, set, nil)

	t.Run("Check we have 2 deployments", func(t *testing.T) {
		result := GetDeploymentsConfig()
		assert.Len(t, result.Deployments, 2)
	})

	t.Run("Check current target is 'remoteserver'", func(t *testing.T) {
		result := FindTargetDeployment()
		assert.Equal(t, "remoteserver", result.ID)
	})

	t.Run("Remove the remoteserver deployment", func(t *testing.T) {
		RemoveDeploymentFromList(c)
		result := GetDeploymentsConfig()
		assert.Len(t, result.Deployments, 1)
	})

	t.Run("Check target reverts back to local", func(t *testing.T) {
		result := FindTargetDeployment()
		assert.Equal(t, "local", result.ID)
	})
}
