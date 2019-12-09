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
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// DeployPerformance takes in a `codewind` object and deploys Codewind and the performance dashboard into the specified namespace
func DeployPerformance(clientset *kubernetes.Clientset, codewind Codewind, deployOptions *DeployOptions) error {

	// Deploy the Performance dashboard
	performanceService := generatePerformanceService(codewind)
	performanceDeploy := generatePerformanceDeploy(codewind)

	log.Infoln("Deploying Codewind Performance Dashboard")
	_, err := clientset.CoreV1().Services(deployOptions.Namespace).Create(&performanceService)
	if err != nil {
		log.Errorf("Error: Unable to create Codewind Performance service: %v\n", err)
		return err
	}
	_, err = clientset.AppsV1().Deployments(deployOptions.Namespace).Create(&performanceDeploy)
	if err != nil {
		log.Errorf("Error: Unable to create Codewind Performance deployment: %v\n", err)
		return err
	}
	return nil
}

func generatePerformanceDeploy(codewind Codewind) appsv1.Deployment {
	labels := map[string]string{
		"app":               PerformancePrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	envVars := setPerformanceEnvVars(codewind)
	return generateDeployment(codewind, PerformancePrefix, codewind.PerformanceImage, PerformanceContainerPort, volumes, volumeMounts, envVars, labels, codewind.ServiceAccountName, false)
}

func generatePerformanceService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":               PerformancePrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, PerformancePrefix, PerformanceContainerPort, labels)
}

func setPerformanceEnvVars(codewind Codewind) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "IN_K8",
			Value: "true",
		},
		{
			Name:  "PORTAL_HTTPS",
			Value: "false",
		},
		{
			Name:  "CODEWIND_INGRESS",
			Value: codewind.Ingress,
		},
	}
}
