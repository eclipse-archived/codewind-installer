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
	"time"

	logr "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ExistingDeployment describes a remote installation of Codewind
type ExistingDeployment struct {
	WorkspaceID       string `json:"workspaceID"`
	Namespace         string `json:"namespace"`
	CodewindURL       string `json:"codewindURL"`
	Version           string `json:"codewindVersion"`
	InstallDate       string `json:"installTime"`
	CodewindAuthRealm string `json:"codewindAuthRealm"`
}

// K8sAPI is the k8s client called by the function
type K8sAPI struct {
	clientset kubernetes.Interface
}

// GetExistingDeployments returns information about the remote installations of codewind, across all namespaces by default
func GetExistingDeployments(namespace string) ([]ExistingDeployment, *RemInstError) {
	config, err := getKubeConfig()

	if err != nil {
		logr.Infof("Unable to retrieve Kubernetes Config %v\n", err)
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	client := K8sAPI{}
	client.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	deployments, RemInstErr := client.findDeployments(namespace)
	if RemInstErr != nil {
		return nil, RemInstErr
	}

	return deployments, nil
}

func (client K8sAPI) findDeployments(namespace string) ([]ExistingDeployment, *RemInstError) {
	deployments, err := client.clientset.AppsV1().Deployments(namespace).List(v1.ListOptions{
		LabelSelector: "app=codewind-pfe",
	})
	if err != nil {
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	var RemoteInstalls []ExistingDeployment
	for _, deployment := range deployments.Items {
		installTime := deployment.GetCreationTimestamp().Format(time.RFC1123)
		var keycloakAddress, cwVersion, authRealm string
		// ensure there are containers in the list, to avoid index errors
		if containers := deployment.Spec.Template.Spec.Containers; len(containers) > 0 {
			env := containers[0].Env
			for _, e := range env {
				if e.Name == "CODEWIND_AUTH_HOST" {
					keycloakAddress = "https://" + e.Value
				}
				if e.Name == "CODEWIND_VERSION" {
					cwVersion = e.Value
				}
				if e.Name == "CODEWIND_AUTH_REALM" {
					authRealm = e.Value
				}
			}
		}

		deployInfo := ExistingDeployment{
			Namespace:         deployment.GetNamespace(),
			WorkspaceID:       deployment.GetLabels()["codewindWorkspace"],
			CodewindURL:       keycloakAddress,
			CodewindAuthRealm: authRealm,
			Version:           cwVersion,
			InstallDate:       installTime,
		}
		RemoteInstalls = append(RemoteInstalls, deployInfo)
	}

	return RemoteInstalls, nil
}
