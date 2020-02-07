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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// DeployPFE : Deploy PFE instance
func DeployPFE(config *restclient.Config, clientset *kubernetes.Clientset, codewindInstance Codewind, deployOptions *DeployOptions) error {

	codewindRoleBindingName := CodewindRoleBindingNamePrefix + "-" + codewindInstance.WorkspaceID
	codewindRoles := CreateCodewindRoles(deployOptions)
	codewindRoleBindings := CreateCodewindRoleBindings(codewindInstance, deployOptions, codewindRoleBindingName)

	codewindTektonClusterRoleBindingName := CodewindTektonClusterRoleBindingName + "-" + codewindInstance.WorkspaceID
	codewindTektonRoles := CreateCodewindTektonClusterRoles(deployOptions)
	codewindTektonRoleBindings := CreateCodewindTektonClusterRoleBindings(codewindInstance, deployOptions, codewindTektonClusterRoleBindingName)

	service := generatePFEService(codewindInstance)
	deploy := generatePFEDeploy(codewindInstance, deployOptions)

	logr.Infof("Checking if '%v' cluster access roles are installed\n", CodewindRolesName)
	clusterRole, err := clientset.RbacV1().ClusterRoles().Get(CodewindRolesName, metav1.GetOptions{})
	if clusterRole != nil && err == nil {
		logr.Infof("Cluster roles '%v' already installed - updating\n", CodewindRolesName)
		_, err = clientset.RbacV1().ClusterRoles().Update(&codewindRoles)
		if err != nil {
			logr.Errorf("Unable to update `%v` cluster access roles: %v\n", CodewindRolesName, err)
			return err
		}
		logr.Infof("Cluster roles '%v' updated complete\n", CodewindRolesName)
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

	logr.Infof("Checking if '%v' Tekton cluster access roles are installed\n", CodewindTektonClusterRolesName)
	tektonclusterRole, err := clientset.RbacV1().ClusterRoles().Get(CodewindTektonClusterRolesName, metav1.GetOptions{})
	if tektonclusterRole != nil && err == nil {
		logr.Infof("Cluster roles '%v' already installed - updating\n", CodewindTektonClusterRolesName)
		_, err = clientset.RbacV1().ClusterRoles().Update(&codewindTektonRoles)
		if err != nil {
			logr.Errorf("Unable to update `%v` Tekton cluster access roles: %v\n", CodewindTektonClusterRolesName, err)
			return err
		}
		logr.Infof("Cluster roles '%v' updated complete\n", CodewindTektonClusterRolesName)
	} else {
		logr.Infof("Adding new '%v' cluster access roles\n", CodewindTektonClusterRolesName)
		_, err = clientset.RbacV1().ClusterRoles().Create(&codewindTektonRoles)
		if err != nil {
			logr.Errorf("Unable to add %v Tekton cluster access roles: %v\n", CodewindTektonClusterRolesName, err)
			return err
		}
	}

	logr.Infof("Checking if '%v' role bindings exist\n", codewindTektonClusterRoleBindingName)
	clusterrolebindings, err := clientset.RbacV1().ClusterRoleBindings().Get(codewindTektonClusterRoleBindingName, metav1.GetOptions{})
	if clusterrolebindings != nil && err == nil {
		logr.Warnf("Cluster Role binding '%v' already exist.\n", codewindTektonClusterRoleBindingName)
	} else {
		logr.Infof("Adding '%v' role binding\n", codewindTektonClusterRoleBindingName)
		_, err = clientset.RbacV1().ClusterRoleBindings().Create(&codewindTektonRoleBindings)
		if err != nil {
			logr.Errorf("Unable to add '%v' access roles: %v\n", codewindTektonClusterRoleBindingName, err)
			return err
		}
	}

	// Determine if we're running on OpenShift on IKS (and thus need to use the ibm-file-bronze storage class)
	storageClass := ""
	sc, err := clientset.StorageV1().StorageClasses().Get(ROKSStorageClass, metav1.GetOptions{})
	if err == nil && sc != nil {
		storageClass = sc.Name
		logr.Infof("Setting storage class to %s\n", storageClass)
	}

	logr.Infof("Creating and setting Codewind PVC %v to %v ", codewindInstance.PVCName, deployOptions.CodewindPVCSize)
	codewindWorkspacePVC := generateCodewindPVC(codewindInstance, deployOptions, storageClass)
	_, err = clientset.CoreV1().PersistentVolumeClaims(deployOptions.Namespace).Create(&codewindWorkspacePVC)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind PVC: %v\n", err)
		return err
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

// generatePFEDeploy : creates a Kubernetes deploy resource
func generatePFEDeploy(codewind Codewind, deployOptions *DeployOptions) appsv1.Deployment {
	labels := map[string]string{
		"app":               PFEPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	volumes, volumeMounts := setPFEVolumes(codewind)
	envVars := setPFEEnvVars(codewind, deployOptions)
	return generateDeployment(codewind, PFEPrefix, codewind.PFEImage, PFEContainerPort, volumes, volumeMounts, envVars, labels, codewind.ServiceAccountName, true)
}

// generatePFEService : creates a Kubernetes service
func generatePFEService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":               PFEPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, PFEPrefix, PFEContainerPort, labels)
}

func setPFEEnvVars(codewind Codewind, deployOptions *DeployOptions) []corev1.EnvVar {

	authHost := deployOptions.KeycloakHost
	if authHost == "" {
		authHost = KeycloakPrefix + codewind.Ingress
	}

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
			Value: GatekeeperPrefix + codewind.Ingress,
		},
		{
			Name:  "INGRESS_PREFIX",
			Value: codewind.Namespace + "." + codewind.RequestedIngress, // provides access to project containers
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
			Value: authHost,
		},
		{
			Name:  "LOG_LEVEL",
			Value: deployOptions.LogLevel,
		},
	}
}

func generateCodewindPVC(codewind Codewind, deployOptions *DeployOptions, storageClass string) corev1.PersistentVolumeClaim {

	labels := map[string]string{
		"app":               PFEPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	pvc := corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   codewind.PVCName,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteMany",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(deployOptions.CodewindPVCSize),
				},
			},
		},
	}

	// If a storage class was passed in, set it in the PVC
	if storageClass != "" {
		pvc.Spec.StorageClassName = &storageClass
	}

	return pvc
}

// setPFEVolumes returns the 2 volumes & corresponding volume mounts required by the PFE container:
// project workspace, buildah volume
func setPFEVolumes(codewind Codewind) ([]corev1.Volume, []corev1.VolumeMount) {

	volumes := []corev1.Volume{
		{
			Name: "shared-workspace",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: codewind.PVCName,
				},
			},
		},
		{
			Name: "buildah-volume",
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "shared-workspace",
			MountPath: "/codewind-workspace",
			SubPath:   codewind.WorkspaceID + "/projects",
		},
		{
			Name:      "buildah-volume",
			MountPath: "/var/lib/containers",
		},
	}

	return volumes, volumeMounts
}
