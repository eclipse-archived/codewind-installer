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

package kube

import (
	"os"

	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubeClientConfig retrieves the Kubernetes client config from the cluster
func GetKubeClientConfig() clientcmd.ClientConfig {
	// Retrieve the Kube client config
	clientconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	return clientconfig
}

// GetCurrentNamespace gets the current namespace in the Kubernetes context
func GetCurrentNamespace() string {
	// Instantiate loader for kubeconfig file.
	kubeconfig := GetKubeClientConfig()
	namespace, _, err := kubeconfig.Namespace()
	if err != nil {
		panic(err)
	}
	return namespace
}

// DetectOpenShift determines if we're running on an OpenShift cluster
// From https://github.com/eclipse/che-operator/blob/2f639261d8b5416b2934591e12925ee0935814dd/pkg/util/util.go#L63
func DetectOpenShift(config *rest.Config) bool {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Errorf("Unable to detect if running on OpenShift: %v\n", err)
		os.Exit(1)
	}
	apiList, err := discoveryClient.ServerGroups()
	if err != nil {
		log.Errorf("Error attempting to retrieve list of API Groups: %v\n", err)
		os.Exit(1)
	}
	apiGroups := apiList.Groups
	for _, group := range apiGroups {
		if group.Name == "route.openshift.io" {
			return true
		}
	}
	return false
}
