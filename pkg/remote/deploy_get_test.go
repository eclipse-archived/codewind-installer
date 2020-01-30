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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

type MockDeploymentOptions struct {
	Namespace         string
	CreationTimestamp time.Time
	Labels            map[string]string
	Env               []corev1.EnvVar
}

func newTestSimpleK8s() *K8sAPI {
	client := K8sAPI{}
	client.clientset = fake.NewSimpleClientset()
	return &client
}

func newTestK8s() *K8sAPI {
	client := K8sAPI{
		clientset: &fake.Clientset{},
	}
	return &client
}

func generateMockDeployment(options MockDeploymentOptions) v1.Deployment {
	return v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         options.Namespace,
			CreationTimestamp: metav1.NewTime(options.CreationTimestamp),
			Labels:            options.Labels,
		},
		Spec: v1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Env: options.Env,
						},
					},
				},
			},
		},
	}
}

func TestSuccessfulDeployGet(t *testing.T) {
	timeNow := time.Now()
	mockDeploymentOptions1 := MockDeploymentOptions{
		Namespace:         "test1",
		CreationTimestamp: timeNow,
		Labels:            map[string]string{"app": "codewind-pfe", "codewindWorkspace": "WID1"},
		Env: []corev1.EnvVar{
			{
				Name:  "CODEWIND_AUTH_HOST",
				Value: "codewind-keycloak-WID1.nip.io",
			},
			{
				Name:  "CODEWIND_VERSION",
				Value: "0.7.0",
			},
			{
				Name:  "CODEWIND_AUTH_REALM",
				Value: "codewind",
			},
		},
	}

	mockDeploymentOptions2 := MockDeploymentOptions{
		Namespace:         "test2",
		CreationTimestamp: timeNow,
		Labels:            map[string]string{"app": "codewind-pfe", "codewindWorkspace": "WID2"},
		Env: []corev1.EnvVar{
			{
				Name:  "CODEWIND_AUTH_HOST",
				Value: "codewind-keycloak-WID2.nip.io",
			},
			{
				Name:  "CODEWIND_VERSION",
				Value: "0.7.0",
			},
			{
				Name:  "CODEWIND_AUTH_REALM",
				Value: "codewind",
			},
		},
	}

	ExistingDeployment1 := ExistingDeployment{
		Namespace:         "test1",
		WorkspaceID:       "WID1",
		CodewindURL:       "https://codewind-keycloak-WID1.nip.io",
		CodewindAuthRealm: "codewind",
		InstallDate:       timeNow.Format(time.RFC1123),
		Version:           "0.7.0",
	}

	ExistingDeployment2 := ExistingDeployment{
		Namespace:         "test2",
		WorkspaceID:       "WID2",
		CodewindURL:       "https://codewind-keycloak-WID2.nip.io",
		CodewindAuthRealm: "codewind",
		InstallDate:       timeNow.Format(time.RFC1123),
		Version:           "0.7.0",
	}

	tests := map[string]struct {
		mockDeploymentOptions []MockDeploymentOptions
		inNamespace           string
		wantedDeploymentInfo  []ExistingDeployment
	}{
		"1 namespace, all deployment options": {
			mockDeploymentOptions: []MockDeploymentOptions{mockDeploymentOptions1},
			inNamespace:           "",
			wantedDeploymentInfo:  []ExistingDeployment{ExistingDeployment1},
		},
		"2 namespaces, all deployment options, search all namespaces": {
			mockDeploymentOptions: []MockDeploymentOptions{mockDeploymentOptions1, mockDeploymentOptions2},
			inNamespace:           "",
			wantedDeploymentInfo:  []ExistingDeployment{ExistingDeployment1, ExistingDeployment2},
		},
		"2 namespaces, all deployment options, search 1 namespace": {
			mockDeploymentOptions: []MockDeploymentOptions{mockDeploymentOptions1, mockDeploymentOptions2},
			inNamespace:           mockDeploymentOptions1.Namespace,
			wantedDeploymentInfo:  []ExistingDeployment{ExistingDeployment1},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			k8s := newTestSimpleK8s()
			deploymentList := []v1.Deployment{}
			for _, options := range test.mockDeploymentOptions {
				deploymentList = append(deploymentList, generateMockDeployment(options))
			}
			fakeDeploymentList := v1.DeploymentList{
				Items: deploymentList,
			}

			// this mocks the k8s client to return the given deployment list
			k8s.clientset = fake.NewSimpleClientset(&fakeDeploymentList)
			got, err := k8s.findDeployments(test.inNamespace)
			checkDeployments(t, got, test.wantedDeploymentInfo, err)
		})
	}
}

func checkDeployments(t *testing.T, got, want []ExistingDeployment, err *RemInstError) {
	t.Helper()
	if err != nil {
		t.Errorf("findDeployments() returned the error: %v", err)
	}
	assert.EqualValues(t, got, want)
}

func Test_Deploy_Get_Error(t *testing.T) {
	k8s := newTestK8s()
	errorDesc := "Mock error getting deployments"
	mockErr := errors.New(errorDesc)
	// this mocks the k8s client to return an error
	k8s.clientset.(*fake.Clientset).Fake.AddReactor("list", "deployments", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1.DeploymentList{}, mockErr
	})
	wantedErr := RemInstError{errOpNotFound, mockErr, errorDesc}
	_, gotErr := k8s.findDeployments("")
	assert.Equal(t, wantedErr, *gotErr)
}
