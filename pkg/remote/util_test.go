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

package remote

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeEnv struct {
	values map[string]string
}

func (f FakeEnv) Getenv(key string) string {
	return f.values[key]
}

func (f FakeEnv) Setenv(key string, value string) {
	f.values[key] = value
}

func NewFakeEnv() FakeEnv {
	f := FakeEnv{}
	f.values = make(map[string]string)
	return f
}

func TestGetImages(t *testing.T) {
	envAllSet := FakeEnv{
		values: map[string]string{
			"PFE_IMAGE":         "/test-pfe",
			"PERFORMANCE_IMAGE": "/test-performance ",
			"KEYCLOAK_IMAGE":    "/test-keycloak ",
			"GATEKEEPER_IMAGE":  "/test-gatekeeper ",
			"PFE_TAG":           "latest",
			"PERFORMANCE_TAG":   "latest",
			"KEYCLOAK_TAG":      "latest",
			"GATEKEEPER_TAG":    "latest",
		},
	}

	noEnvVarsSet := FakeEnv{
		values: map[string]string{},
	}

	t.Run("success case - all images set", func(t *testing.T) {
		pfeImage, perfImage, keycloakImage, gatekeeperImage := GetImages(envAllSet)
		values := envAllSet.values
		expectedPfeImage := values["PFE_IMAGE"] + ":" + values["PFE_TAG"]
		expectedPerfImage := values["PERFORMANCE_IMAGE"] + ":" + values["PERFORMANCE_TAG"]
		expectedKeycloakImage := values["KEYCLOAK_IMAGE"] + ":" + values["KEYCLOAK_TAG"]
		expectedGatekeeperImage := values["GATEKEEPER_IMAGE"] + ":" + values["GATEKEEPER_TAG"]
		assert.Equal(t, expectedPfeImage, pfeImage)
		assert.Equal(t, expectedPerfImage, perfImage)
		assert.Equal(t, expectedKeycloakImage, keycloakImage)
		assert.Equal(t, expectedGatekeeperImage, gatekeeperImage)
	})

	t.Run("success case - no env vars set, uses defaults", func(t *testing.T) {
		pfeImage, perfImage, keycloakImage, gatekeeperImage := GetImages(noEnvVarsSet)
		expectedPfeImage := PFEImage + ":" + PFEImageTag
		expectedPerfImage := PerformanceImage + ":" + PerformanceTag
		expectedKeycloakImage := KeycloakImage + ":" + KeycloakImageTag
		expectedGatekeeperImage := GatekeeperImage + ":" + GatekeeperImageTag
		assert.Equal(t, expectedPfeImage, pfeImage)
		assert.Equal(t, expectedPerfImage, perfImage)
		assert.Equal(t, expectedKeycloakImage, keycloakImage)
		assert.Equal(t, expectedGatekeeperImage, gatekeeperImage)
	})

}
