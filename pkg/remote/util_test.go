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

package remote

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setTestEnvVars(t *testing.T, testEnvVars map[string]string) func() {
	t.Helper()

	// save the current env vars, to reset after the test is run
	currentEnvVars := map[string]string{}
	for envVar := range testEnvVars {
		currentEnvVars[envVar] = os.Getenv(envVar)
	}

	// set env vars to test given values
	for envVar, value := range testEnvVars {
		os.Setenv(envVar, value)
	}

	resetEnvVars := func() {
		for envVar, value := range currentEnvVars {
			os.Setenv(envVar, value)
		}
	}

	return resetEnvVars
}

func TestGetImages(t *testing.T) {
	envAllSet := map[string]string{
		"PFE_IMAGE":         "/test-pfe",
		"PERFORMANCE_IMAGE": "/test-performance",
		"KEYCLOAK_IMAGE":    "/test-keycloak",
		"GATEKEEPER_IMAGE":  "/test-gatekeeper",
		"PFE_TAG":           "latest",
		"PERFORMANCE_TAG":   "latest",
		"KEYCLOAK_TAG":      "latest",
		"GATEKEEPER_TAG":    "latest",
	}

	t.Run("success case - all images set", func(t *testing.T) {
		resetEnvVars := setTestEnvVars(t, envAllSet)
		defer resetEnvVars()

		pfeImage, perfImage, keycloakImage, gatekeeperImage := GetImages()
		expectedPfeImage := envAllSet["PFE_IMAGE"] + ":" + envAllSet["PFE_TAG"]
		expectedPerfImage := envAllSet["PERFORMANCE_IMAGE"] + ":" + envAllSet["PERFORMANCE_TAG"]
		expectedKeycloakImage := envAllSet["KEYCLOAK_IMAGE"] + ":" + envAllSet["KEYCLOAK_TAG"]
		expectedGatekeeperImage := envAllSet["GATEKEEPER_IMAGE"] + ":" + envAllSet["GATEKEEPER_TAG"]
		assert.Equal(t, expectedPfeImage, pfeImage)
		assert.Equal(t, expectedPerfImage, perfImage)
		assert.Equal(t, expectedKeycloakImage, keycloakImage)
		assert.Equal(t, expectedGatekeeperImage, gatekeeperImage)
	})

	t.Run("success case - no env vars set, uses defaults", func(t *testing.T) {
		resetEnvVars := setTestEnvVars(t, map[string]string{})
		defer resetEnvVars()

		pfeImage, perfImage, keycloakImage, gatekeeperImage := GetImages()
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

type testParamaterOptions = struct {
	name               string
	image              string
	port               int
	secrets            map[string]string
	labels             map[string]string
	volumes            []corev1.Volume
	volumeMounts       []corev1.VolumeMount
	envVars            []corev1.EnvVar
	serviceAccountName string
	privileged         bool
}

var defaultParams = testParamaterOptions{
	name:               "name",
	image:              "imagename",
	port:               200,
	secrets:            map[string]string{},
	labels:             map[string]string{},
	volumes:            []corev1.Volume{},
	volumeMounts:       []corev1.VolumeMount{},
	envVars:            []corev1.EnvVar{},
	serviceAccountName: "sac-name",
	privileged:         false,
}

func TestGenerateDeployment(t *testing.T) {
	t.Run("success case - returns correct deployment", func(t *testing.T) {
		replicas := int32(1)
		deployment := generateDeployment(MockCodewind, defaultParams.name, defaultParams.image, defaultParams.port, defaultParams.volumes, defaultParams.volumeMounts, defaultParams.envVars, defaultParams.labels, defaultParams.serviceAccountName, defaultParams.privileged)
		expectedDeployment := appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultParams.name + "-" + MockCodewind.WorkspaceID,
				Namespace: MockCodewind.Namespace,
				Labels:    defaultParams.labels,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: defaultParams.labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: defaultParams.labels,
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: defaultParams.serviceAccountName,
						Volumes:            []corev1.Volume{},
						Containers: []corev1.Container{
							{
								Name:            defaultParams.name,
								Image:           defaultParams.image,
								ImagePullPolicy: ImagePullPolicy,
								SecurityContext: &corev1.SecurityContext{
									Privileged: &defaultParams.privileged,
								},
								VolumeMounts: defaultParams.volumeMounts,
								Env:          defaultParams.envVars,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: int32(defaultParams.port),
									},
								},
							},
						},
					},
				},
			},
		}
		assert.Equal(t, expectedDeployment, deployment)
	})
}

func TestGenerateSecrets(t *testing.T) {
	t.Run("success case - returns generated secrets", func(t *testing.T) {
		secrets := generateSecrets(MockCodewind, defaultParams.name, defaultParams.secrets, defaultParams.labels)
		expectedSecrets := corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultParams.name + "-" + MockCodewind.WorkspaceID,
				Namespace: MockCodewind.Namespace,
				Labels:    defaultParams.labels,
			},
			StringData: defaultParams.secrets,
		}
		assert.Equal(t, expectedSecrets, secrets)
	})
}

func TestGenerateService(t *testing.T) {
	t.Run("success case - returns generated service", func(t *testing.T) {
		service := generateService(MockCodewind, defaultParams.name, defaultParams.port, defaultParams.labels)
		expectedService := corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultParams.name + "-" + MockCodewind.WorkspaceID,
				Namespace: MockCodewind.Namespace,
				Labels:    defaultParams.labels,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Port: int32(defaultParams.port),
						Name: defaultParams.name + "-http",
					},
				},
				Selector: defaultParams.labels,
			},
		}
		assert.Equal(t, expectedService, service)
	})
}
