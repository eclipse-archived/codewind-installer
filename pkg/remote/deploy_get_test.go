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
	"fmt"
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

func newTestSimpleK8s() *KubernetesAPI {
	client := KubernetesAPI{}
	client.clientset = fake.NewSimpleClientset()
	return &client
}

func newTestK8s() *KubernetesAPI {
	client := KubernetesAPI{
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

func Test_Deploy_Get(t *testing.T) {
	k8s := newTestSimpleK8s()
	timeNow := time.Now()

	deploy1 := generateMockDeployment(MockDeploymentOptions{
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
	})

	deploy2 := generateMockDeployment(MockDeploymentOptions{
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
	})

	// mock the deployment list returned from the k8s client
	fakeDeploymentList := v1.DeploymentList{
		Items: []v1.Deployment{
			deploy1, deploy2,
		},
	}
	k8s.clientset = fake.NewSimpleClientset(&fakeDeploymentList)

	ExistingDeployment1 := ExistingDeployment{
		Namespace:         "test1",
		WorkspaceID:       "WID1",
		CodewindURL:       "https://codewind-keycloak-WID1.nip.io",
		CodewindAuthRealm: "codewind",
		InstallDate:       timeNow.Format("02-Jan-2006"),
		Version:           "0.7.0",
	}

	ExistingDeployment2 := ExistingDeployment{
		Namespace:         "test2",
		WorkspaceID:       "WID2",
		CodewindURL:       "https://codewind-keycloak-WID2.nip.io",
		CodewindAuthRealm: "codewind",
		InstallDate:       timeNow.Format("02-Jan-2006"),
		Version:           "0.7.0",
	}

	// using an empty string searches all namespaces
	gotForAllNamespaces, err := k8s.FindDeployments("")
	wantForAllNamespaces := []ExistingDeployment{
		ExistingDeployment1, ExistingDeployment2,
	}
	if err != nil {
		fmt.Println(err)
		t.Fatal("findDeployments() should not return an error")
	}
	assert.EqualValues(t, wantForAllNamespaces, gotForAllNamespaces)

	// now search a particular namespace, and check the details for only one is returned
	wantForSingleNamespace := []ExistingDeployment{
		ExistingDeployment1,
	}

	gotForSingleNamespace, err := k8s.FindDeployments("test1")
	if err != nil {
		fmt.Println(err)
		t.Fatal("findDeployments() should not return an error")
	}
	assert.Equal(t, wantForSingleNamespace, gotForSingleNamespace)
}

func Test_Deploy_Get_Error(t *testing.T) {
	k8s := newTestK8s()
	errorDesc := "Mock error getting deployments"
	mockErr := errors.New(errorDesc)
	// this mocks the kube client to return an error
	k8s.clientset.(*fake.Clientset).Fake.AddReactor("list", "deployments", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1.DeploymentList{}, mockErr
	})
	wantedErr := RemInstError{errOpNotFound, mockErr, errorDesc}
	_, gotErr := k8s.FindDeployments("")
	assert.Equal(t, wantedErr, *gotErr)
}
