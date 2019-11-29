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
	"strconv"

	"github.com/eclipse/codewind-installer/pkg/appconstants"
	logr "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// DeployPFE : Deploy PFE instance
func DeployPFE(config *restclient.Config, clientset *kubernetes.Clientset, codewindInstance Codewind, deployOptions *DeployOptions) error {

	codewindRoleBindingName := CodewindRoleBindingNamePrefix + "-" + codewindInstance.WorkspaceID

	codewindRoles := CreateCodewindRoles(deployOptions)
	codewindRoleBindings := CreateCodewindRoleBindings(codewindInstance, deployOptions, codewindRoleBindingName)

	service := createPFEService(codewindInstance)
	deploy := createPFEDeploy(codewindInstance, deployOptions)

	logr.Infof("Checking if '%v' cluster access roles are installed\n", CodewindRolesName)
	clusterRole, err := clientset.RbacV1().ClusterRoles().Get(CodewindRolesName, metav1.GetOptions{})
	if clusterRole != nil && err == nil {
		logr.Infof("Cluster roles '%v' already installed\n", CodewindRolesName)
	} else {
		logr.Infof("Adding new '%v' cluster access roles\n", CodewindRolesName)
		_, err = clientset.RbacV1().ClusterRoles().Create(&codewindRoles)
		if err != nil {
			logr.Errorf("Unable to add %v cluster access roles: %v\n", CodewindRolesName, err)
			return err
		}
	}

	logr.Infof("Checking if '%v' role bindings exist\n", codewindRoleBindingName)
	rolebindings, err := clientset.RbacV1().RoleBindings(codewindInstance.Namespace).Get(codewindRoleBindingName, metav1.GetOptions{})
	if rolebindings != nil && err == nil {
		logr.Warnf("Role binding '%v' already exist.\n", codewindRoleBindingName)
	} else {
		logr.Infof("Adding '%v' role binding\n", codewindRoleBindingName)
		_, err = clientset.RbacV1().RoleBindings(codewindInstance.Namespace).Create(&codewindRoleBindings)
		if err != nil {
			logr.Errorf("Unable to add '%v' access roles: %v\n", codewindRoleBindingName, err)
			return err
		}
	}

	logr.Infoln("Deploying Codewind Service")
	_, err = clientset.CoreV1().Services(deployOptions.Namespace).Create(&service)
	if err != nil {
		logr.Errorf("Unable to create Codewind service: %v\n", err)
		return err
	}
	_, err = clientset.AppsV1().Deployments(deployOptions.Namespace).Create(&deploy)
	if err != nil {
		logr.Errorf("Unable to create Codewind deployment: %v\n", err)
		return err
	}
	return nil
}

// createPFEDeploy : creates a Kubernetes deploy resource
func createPFEDeploy(codewind Codewind, deployOptions *DeployOptions) appsv1.Deployment {
	labels := map[string]string{
		"app":               PFEPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	volumes, volumeMounts := setPFEVolumes(codewind)
	envVars := setPFEEnvVars(codewind, deployOptions)
	return generateDeployment(codewind, PFEPrefix, codewind.PFEImage, PFEContainerPort, volumes, volumeMounts, envVars, labels)
}

// createPFEService : creates a Kubernetes service
func createPFEService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":               PFEPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, PFEPrefix, PFEContainerPort, labels)
}

func setPFEEnvVars(codewind Codewind, deployOptions *DeployOptions) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "TEKTON_PIPELINE",
			Value: "tekton-pipelines",
		},
		{
			Name:  "IN_K8",
			Value: "true",
		},
		{
			Name:  "PORTAL_HTTPS",
			Value: "true",
		},
		{
			Name:  "KUBE_NAMESPACE",
			Value: codewind.Namespace,
		},
		{
			Name:  "TILLER_NAMESPACE",
			Value: codewind.Namespace,
		},
		{
			Name:  "CHE_WORKSPACE_ID",
			Value: codewind.WorkspaceID,
		},
		{
			Name:  "PVC_NAME",
			Value: codewind.PVCName,
		},
		{
			Name:  "SERVICE_NAME",
			Value: "codewind-" + codewind.WorkspaceID,
		},
		{
			Name:  "SERVICE_ACCOUNT_NAME",
			Value: codewind.ServiceAccountName,
		},
		{
			Name:  "MICROCLIMATE_RELEASE_NAME",
			Value: "RELEASE-NAME",
		},
		{
			Name:  "HOST_WORKSPACE_DIRECTORY",
			Value: "/projects",
		},
		{
			Name:  "CONTAINER_WORKSPACE_DIRECTORY",
			Value: "/codewind-workspace",
		},
		{
			Name:  "CODEWIND_VERSION",
			Value: appconstants.VersionNum,
		},
		{
			Name:  "OWNER_REF_NAME",
			Value: codewind.OwnerReferenceName,
		},
		{
			Name:  "OWNER_REF_UID",
			Value: string(codewind.OwnerReferenceUID),
		},
		{
			Name:  "CODEWIND_PERFORMANCE_SERVICE",
			Value: PerformancePrefix + "-" + codewind.WorkspaceID,
		},
		{
			Name:  "CHE_INGRESS_HOST",
			Value: codewind.Ingress,
		},
		{
			Name:  "INGRESS_PREFIX",
			Value: codewind.RequestedIngress, // provides access to project containers
		},
		{
			Name:  "ON_OPENSHIFT",
			Value: strconv.FormatBool(codewind.OnOpenShift),
		},
		{
			Name:  "CODEWIND_AUTH_REALM",
			Value: deployOptions.KeycloakRealm,
		},
		{
			Name:  "CODEWIND_AUTH_HOST",
			Value: KeycloakPrefix + codewind.Ingress,
		},
	}
}

// setPFEVolumes returns the 3 volumes & corresponding volume mounts required by the PFE container:
// project workspace, buildah volume, and the docker registry secret (the latter of which is optional)
func setPFEVolumes(codewind Codewind) ([]corev1.Volume, []corev1.VolumeMount) {
	secretMode := int32(511)
	isOptional := true

	volumes := []corev1.Volume{
		// {
		// 	Name: "shared-workspace",
		// 	VolumeSource: corev1.VolumeSource{
		// 		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
		// 			ClaimName: codewind.PVCName,
		// 		},
		// 	},
		// },
		{
			Name: "buildah-volume",
		},
		{
			Name: "registry-secret",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &secretMode,
					SecretName:  codewind.PullSecret,
					Optional:    &isOptional,
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		// {
		// 	Name:      "shared-workspace",
		// 	MountPath: "/codewind-workspace",
		// 	SubPath:   codewind.WorkspaceID + "/projects",
		// },
		{
			Name:      "buildah-volume",
			MountPath: "/var/lib/containers",
		},
		{
			Name:      "registry-secret",
			MountPath: "/tmp/secret",
		},
	}

	return volumes, volumeMounts
}
