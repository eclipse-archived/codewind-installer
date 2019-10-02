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

func Test_GetActiveDeployment(t *testing.T) {
	InitDeploymentConfigIfRequired()
	t.Run("ActiveDeployment", func(t *testing.T) {
		ResetDeploymentsFile()
		result := FindTargetDeployment()
		assert.Equal(t, "local", result.Name)
		assert.Equal(t, "Codewind local deployment", result.Label)
		assert.Equal(t, "", result.URL)
	})
}

func Test_GetDeploymentsConfig(t *testing.T) {
	t.Run("GetDeploymentsConfig", func(t *testing.T) {
		ResetDeploymentsFile()
		result := GetDeploymentsConfig()
		assert.Equal(t, "local", result.Active)
		assert.Len(t, result.Deployments, 1)
	})
}

//TODO:  add coverage
