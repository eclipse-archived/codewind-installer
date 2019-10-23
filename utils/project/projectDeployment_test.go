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
const testDeploymentID = "local"

// Test_ProjectDeployment :  Tests
func Test_ProjectDeployment(t *testing.T) {
	ResetTargetFile(testProjectID)

	t.Run("Asserts there are no target deployments", func(t *testing.T) {
		deploymentTargets, projError := ListTargetDeployments(testProjectID)
		if projError != nil {
			t.Fail()
		}
		assert.Len(t, deploymentTargets.DeploymentTargets, 0)
	})

	t.Run("Add project to local deployment", func(t *testing.T) {
		projError := AddDeploymentTarget(testProjectID, testDeploymentID)
		if projError != nil {
			t.Fail()
		}
	})

	t.Run("Asserts re-adding the same deployment fails", func(t *testing.T) {
		projError := AddDeploymentTarget(testProjectID, testDeploymentID)
		if projError == nil {
			t.Fail()
		}
		assert.Equal(t, errOpConflict, projError.Op)
	})

	t.Run("Asserts there is just 1 target deployment added", func(t *testing.T) {
		deploymentTargets, projError := ListTargetDeployments(testProjectID)
		if projError != nil {
			t.Fail()
		}
		assert.Len(t, deploymentTargets.DeploymentTargets, 1)
	})

	t.Run("Asserts an unknown deployment can not be removed", func(t *testing.T) {
		projError := RemoveDeploymentTarget(testProjectID, "test-AnUnknownDeploymentID")
		if projError == nil {
			t.Fail()
		}
		assert.Equal(t, errOpNotFound, projError.Op)
	})

	t.Run("Asserts removing a known deployment is successful", func(t *testing.T) {
		projError := RemoveDeploymentTarget(testProjectID, "local")
		if projError != nil {
			t.Fail()
		}
	})

	t.Run("Asserts there are no targets left for this project", func(t *testing.T) {
		deploymentTargets, projError := ListTargetDeployments(testProjectID)
		if projError != nil {
			t.Fail()
		}
		assert.Len(t, deploymentTargets.DeploymentTargets, 0)
	})

	t.Run("Asserts attempting to manage an invalid project ID fails", func(t *testing.T) {
		projError := AddDeploymentTarget("bad-project-ID", testDeploymentID)
		if projError == nil {
			t.Fail()
		}
		assert.Equal(t, errOpInvalidID, projError.Op)
	})

}
