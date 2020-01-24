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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func newTestSimpleK8s() *KubernetesAPI {
	client := KubernetesAPI{}
	client.clientset = fake.NewSimpleClientset()
	return &client
}
func Test_Deploy_Get(t *testing.T) {
	k8s := newTestSimpleK8s()
	timeNow := time.Now()

	// Mock the deployment list which is returned from the k8s client
	fakeDeploymentList := v1.DeploymentList{
		Items: []v1.Deployment{
			v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "test1",
					CreationTimestamp: metav1.NewTime(timeNow),
					Labels:            map[string]string{"app": "codewind-pfe", "codewindWorkspace": "WID1"},
				},
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
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
								},
							},
						},
					},
				},
			},
			v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "test2",
					CreationTimestamp: metav1.NewTime(timeNow),
					Labels:            map[string]string{"app": "codewind-pfe", "codewindWorkspace": "WID2"},
				},
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Env: []corev1.EnvVar{
										{
											Name:  "CODEWIND_VERSION",
											Value: "0.7.0",
										},
										{
											Name:  "CODEWIND_AUTH_HOST",
											Value: "codewind-keycloak-WID2.nip.io",
										},
										{
											Name:  "CODEWIND_AUTH_REALM",
											Value: "codewind",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	k8s.clientset = fake.NewSimpleClientset(&fakeDeploymentList)

	// using an empty string searches all namespaces
	gotForAllNamespaces, err := k8s.FindDeployments("")
	wantForAllNamespaces := []ExistingDeployment{
		ExistingDeployment{
			Namespace:         "test1",
			WorkspaceID:       "WID1",
			CodewindURL:       "https://codewind-keycloak-WID1.nip.io",
			CodewindAuthRealm: "codewind",
			InstallDate:       timeNow.Format("02-Jan-2006"),
			Version:           "0.7.0",
		},
		ExistingDeployment{
			Namespace:         "test2",
			WorkspaceID:       "WID2",
			CodewindURL:       "https://codewind-keycloak-WID2.nip.io",
			CodewindAuthRealm: "codewind",
			InstallDate:       timeNow.Format("02-Jan-2006"),
			Version:           "0.7.0",
		},
	}
	if err != nil {
		fmt.Println(err)
		t.Fatal("findDeployments() should not return an error")
	}
	assert.EqualValues(t, wantForAllNamespaces, gotForAllNamespaces)

	// now search a particular namespace, and check the details for only one is returned
	wantSingleNamespace := []ExistingDeployment{
		ExistingDeployment{
			Namespace:         "test1",
			WorkspaceID:       "WID1",
			CodewindURL:       "https://codewind-keycloak-WID1.nip.io",
			CodewindAuthRealm: "codewind",
			InstallDate:       timeNow.Format("02-Jan-2006"),
			Version:           "0.7.0",
		},
	}
	gotSingleNamespace, err := k8s.FindDeployments("test1")
	if err != nil {
		fmt.Println(err)
		t.Fatal("findDeployments() should not return an error")
	}
	assert.EqualValues(t, wantSingleNamespace, gotSingleNamespace)
}
