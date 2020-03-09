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
	"github.com/eclipse/codewind-installer/pkg/remote/kube"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	logr "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// RemoveDeploymentOptions : Deployment removal options
type RemoveDeploymentOptions struct {
	Namespace   string
	WorkspaceID string
}

const (
	// ResourceNotProcessed : Resource not processed
	ResourceNotProcessed = 0
	// ResourceFound : Resource located
	ResourceFound = 1
	// ResourceNotFound : Resource could not be located
	ResourceNotFound = 2
	// ResourceRemoved : Resource was successfully removed
	ResourceRemoved = 3
	// ResourceSkipped : Resource was not removed and was skipped
	ResourceSkipped = 4
	// ResourceRemoveFailed : Resource removal failed
	ResourceRemoveFailed = 5
)

// RemovalResult : Status for each component
type RemovalResult struct {

	// Pods
	StatusPODGatekeeper  int
	StatusPODPFE         int
	StatusPODPerformance int
	StatusPODKeycloak    int

	// Services
	StatusServiceGatekeeper  int
	StatusServicePFE         int
	StatusServicePerformance int
	StatusServiceKeycloak    int

	// Deployments
	StatusDeploymentGatekeeper  int
	StatusDeploymentPFE         int
	StatusDeploymentPerformance int
	StatusDeploymentKeycloak    int

	// Secrets
	StatusSecretsCodewind int
	StatusSecretsKeycloak int

	// Service account
	StatusServiceAccount int

	// Role bindings
	StatusRoleBindings       int
	StatusTektonRoleBindings int

	// Persistent volume claims
	StatusPVCCodewind int
	StatusPVCKeycloak int

	// Ingress/Routes
	StatusIngressGatekeeper int
	StatusIngressKeycloak   int
}

// RemoveRemote : Remove remote install from Kube
func RemoveRemote(remoteRemovalOptions *RemoveDeploymentOptions) (*RemovalResult, *RemInstError) {
	namespace := remoteRemovalOptions.Namespace
	config, err := getKubeConfig()
	if err != nil {
		logr.Infof("Unable to retrieve Kubernetes Config %v\n", err)
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	// Determine if we're running on OpenShift or not.
	onOpenShift := kube.DetectOpenShift(config)
	logr.Infof("Running on openshift: %t\n", onOpenShift)

	removalStatus := RemovalResult{
		StatusPODGatekeeper:         ResourceNotProcessed,
		StatusPODPFE:                ResourceNotProcessed,
		StatusPODPerformance:        ResourceNotProcessed,
		StatusServiceGatekeeper:     ResourceNotProcessed,
		StatusServicePFE:            ResourceNotProcessed,
		StatusServicePerformance:    ResourceNotProcessed,
		StatusDeploymentGatekeeper:  ResourceNotProcessed,
		StatusDeploymentPFE:         ResourceNotProcessed,
		StatusDeploymentPerformance: ResourceNotProcessed,
		StatusSecretsCodewind:       ResourceNotProcessed,
		StatusServiceAccount:        ResourceNotProcessed,
		StatusRoleBindings:          ResourceNotProcessed,
		StatusTektonRoleBindings:    ResourceNotProcessed,
		StatusPVCCodewind:           ResourceNotProcessed,
		StatusIngressGatekeeper:     ResourceNotProcessed,
	}

	if err != nil {
		logr.Infof("Unable to retrieve Kubernetes Config %v\n", err)
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logr.Infof("Unable to retrieve Kubernetes clientset %v\n", err)
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	// Check if namespace exists
	logr.Infof("Checking namespace %v exists\n", namespace)
	_, err = clientset.CoreV1().Namespaces().Get(namespace, v1.GetOptions{})
	if err != nil {
		logr.Errorf("Unable to locate %v namespace: %v", namespace, err)
		return nil, &RemInstError{errOpCreateNamespace, err, err.Error()}
	}
	logr.Infof("Found '%v' namespace\n", namespace)

	logr.Trace("Removing Codewind deployments")
	status, err := deleteDeployment(remoteRemovalOptions, clientset, "app="+PFEPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusDeploymentPFE = status
	status, err = deleteDeployment(remoteRemovalOptions, clientset, "app="+PerformancePrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusDeploymentPerformance = status
	status, err = deleteDeployment(remoteRemovalOptions, clientset, "app="+GatekeeperPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusDeploymentGatekeeper = status

	logr.Trace("Removing Codewind services")
	status, err = deleteService(remoteRemovalOptions, clientset, "app="+PFEPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusServicePFE = status
	status, err = deleteService(remoteRemovalOptions, clientset, "app="+PerformancePrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusServicePerformance = status
	status, err = deleteService(remoteRemovalOptions, clientset, "app="+GatekeeperPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusServiceGatekeeper = status

	logr.Trace("Removing Codewind secrets")
	status, err = deleteSecrets(remoteRemovalOptions, clientset, "app="+GatekeeperPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusSecretsCodewind = status

	logr.Trace("Removing Codewind PVC")
	status, err = deletePVC(remoteRemovalOptions, clientset, "app="+PFEPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusPVCCodewind = status

	logr.Trace("Removing Codewind role bindings")
	status, err = deleteRoleBindings(remoteRemovalOptions, clientset, "codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusRoleBindings = status

	logr.Trace("Removing Codewind Tekton role bindings")
	status, err = deleteTektonClusterRoleBindings(remoteRemovalOptions, clientset, "app="+CodewindTektonClusterRoleBindingName+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusTektonRoleBindings = status

	logr.Trace("Removing Codewind service account")
	status, err = deleteServiceAccount(remoteRemovalOptions, clientset, "app=codewind-"+remoteRemovalOptions.WorkspaceID+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusServiceAccount = status

	if onOpenShift {
		logr.Trace("Removing Codewind route")
		status, err = deleteRoute(config, remoteRemovalOptions, clientset, "app="+GatekeeperPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
		removalStatus.StatusIngressGatekeeper = status
	} else {
		logr.Trace("Removing Codewind ingress")
		status, err = deleteIngress(remoteRemovalOptions, clientset, "app="+GatekeeperPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
		removalStatus.StatusIngressGatekeeper = status
	}

	logr.Info("Removal summary:")
	logr.Infof("Codewind PFE Deployment: %v", getStatus(removalStatus.StatusDeploymentPFE))
	logr.Infof("Codewind PFE Service: %v", getStatus(removalStatus.StatusServicePFE))
	logr.Infof("Codewind PFE PVC: %v", getStatus(removalStatus.StatusPVCCodewind))
	logr.Infof("Codewind Performance Deployment: %v", getStatus(removalStatus.StatusDeploymentPerformance))
	logr.Infof("Codewind Performance Service: %v", getStatus(removalStatus.StatusServicePerformance))
	logr.Infof("Codewind Gatekeeper Deployment: %v", getStatus(removalStatus.StatusDeploymentGatekeeper))
	logr.Infof("Codewind Gatekeeper Service: %v", getStatus(removalStatus.StatusServiceGatekeeper))
	logr.Infof("Codewind Gatekeeper Ingress: %v", getStatus(removalStatus.StatusIngressGatekeeper))
	logr.Infof("Codewind Role Bindings: %v", getStatus(removalStatus.StatusRoleBindings))
	logr.Infof("Codewind Tekton Role Bindings: %v", getStatus(removalStatus.StatusTektonRoleBindings))
	logr.Infof("Codewind Service Account: %v", getStatus(removalStatus.StatusServiceAccount))
	logr.Infof("Kubernetes namespace: CWCTL will not remove the namespace automatically, use 'kubectl delete namespace %s' if you would like to remove it", remoteRemovalOptions.Namespace)

	return &removalStatus, nil
}

// RemoveRemoteKeycloak : Remove remote keycloak install from Kube
func RemoveRemoteKeycloak(remoteRemovalOptions *RemoveDeploymentOptions) (*RemovalResult, *RemInstError) {
	namespace := remoteRemovalOptions.Namespace
	config, err := getKubeConfig()
	if err != nil {
		logr.Infof("Unable to retrieve Kubernetes Config %v\n", err)
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	// Determine if we're running on OpenShift or not.
	onOpenShift := kube.DetectOpenShift(config)
	logr.Infof("Running on Openshift: %t\n", onOpenShift)

	removalStatus := RemovalResult{
		StatusPODKeycloak:        ResourceNotProcessed,
		StatusServiceKeycloak:    ResourceNotProcessed,
		StatusDeploymentKeycloak: ResourceNotProcessed,
		StatusSecretsKeycloak:    ResourceNotProcessed,
		StatusServiceAccount:     ResourceNotProcessed,
		StatusPVCKeycloak:        ResourceNotProcessed,
		StatusIngressKeycloak:    ResourceNotProcessed,
	}

	if err != nil {
		logr.Errorf("Unable to retrieve Kubernetes Config %v\n", err)
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logr.Errorf("Unable to retrieve Kubernetes clientset %v\n", err)
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	// Check if namespace exists
	logr.Infof("Checking namespace %v exists\n", namespace)
	_, err = clientset.CoreV1().Namespaces().Get(namespace, v1.GetOptions{})
	if err != nil {
		logr.Errorf("Unable to locate %v namespace: %v", namespace, err)
		return nil, &RemInstError{errOpCreateNamespace, err, err.Error()}
	}
	logr.Infof("Found '%v' namespace\n", namespace)

	logr.Trace("Removing Keycloak deployment")
	status, err := deleteDeployment(remoteRemovalOptions, clientset, "app="+KeycloakPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusDeploymentKeycloak = status

	logr.Trace("Removing Keycloak service")
	status, err = deleteService(remoteRemovalOptions, clientset, "app="+KeycloakPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusServiceKeycloak = status

	logr.Trace("Removing Keycloak secrets")
	status, err = deleteSecrets(remoteRemovalOptions, clientset, "app="+KeycloakPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusSecretsKeycloak = status

	logr.Trace("Removing Keycloak PVC")
	status, err = deletePVC(remoteRemovalOptions, clientset, "app="+KeycloakPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusPVCKeycloak = status

	logr.Trace("Removing Keycloak service account")
	status, err = deleteServiceAccount(remoteRemovalOptions, clientset, "app=keycloak-"+remoteRemovalOptions.WorkspaceID+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
	removalStatus.StatusServiceAccount = status

	if onOpenShift {
		logr.Trace("Removing Keycloak route")
		status, err = deleteRoute(config, remoteRemovalOptions, clientset, "app="+KeycloakPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
		removalStatus.StatusIngressKeycloak = status
	} else {
		logr.Trace("Removing Keycloak ingress")
		status, err = deleteIngress(remoteRemovalOptions, clientset, "app="+KeycloakPrefix+",codewindWorkspace="+remoteRemovalOptions.WorkspaceID)
		removalStatus.StatusIngressKeycloak = status
	}

	logr.Info("Removal summary:")
	logr.Infof("Keycloak Deployment: %v", getStatus(removalStatus.StatusDeploymentKeycloak))
	logr.Infof("Keycloak Service: %v", getStatus(removalStatus.StatusServiceKeycloak))
	logr.Infof("Keycloak PVC: %v", getStatus(removalStatus.StatusPVCKeycloak))
	logr.Infof("Keycloak Ingress: %v", getStatus(removalStatus.StatusIngressKeycloak))
	logr.Infof("Keycloak Secrets: %v", getStatus(removalStatus.StatusSecretsKeycloak))
	logr.Infof("Keycloak Service Account: %v", getStatus(removalStatus.StatusServiceAccount))
	logr.Infof("Kubernetes namespace: CWCTL will not remove the namespace automatically, use 'kubectl delete namespace %s' if you would like to remove it", remoteRemovalOptions.Namespace)
	return &removalStatus, nil
}

func getStatus(status int) string {
	switch status {
	case ResourceNotProcessed:
		return "Not processed"
	case ResourceFound:
		return "Found"
	case ResourceNotFound:
		return "Not found"
	case ResourceRemoved:
		return "Removed"
	case ResourceSkipped:
		return "Skipped"
	case ResourceRemoveFailed:
		return "Removal failed"
	default:
		return ""
	}
}

func deleteDeployment(remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceNotFound
	deploymentList, err := clientset.AppsV1().Deployments(remoteRemovalOptions.Namespace).List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return phase, err
	}
	if deploymentList != nil && deploymentList.Items != nil && len(deploymentList.Items) == 1 {
		phase = ResourceFound
		err := clientset.AppsV1().Deployments(remoteRemovalOptions.Namespace).Delete(deploymentList.Items[0].GetName(), nil)
		if err != nil {
			phase = ResourceRemoveFailed
		} else {
			phase = ResourceRemoved
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}

func deletePod(remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceNotFound
	podList, err := clientset.CoreV1().Pods(remoteRemovalOptions.Namespace).List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return phase, err
	}
	if podList != nil && podList.Items != nil && len(podList.Items) == 1 {
		phase = ResourceFound
		err := clientset.CoreV1().Pods(remoteRemovalOptions.Namespace).Delete(podList.Items[0].GetName(), nil)
		if err != nil {
			phase = ResourceRemoveFailed
		} else {
			phase = ResourceRemoved
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}

func deleteService(remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceNotFound
	serviceList, err := clientset.CoreV1().Services(remoteRemovalOptions.Namespace).List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return phase, err
	}
	if serviceList != nil && serviceList.Items != nil && len(serviceList.Items) == 1 {
		phase = ResourceFound
		err := clientset.CoreV1().Services(remoteRemovalOptions.Namespace).Delete(serviceList.Items[0].GetName(), nil)
		if err != nil {
			phase = ResourceRemoveFailed
		} else {
			phase = ResourceRemoved
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}

func deleteSecrets(remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceNotFound
	secretList, err := clientset.CoreV1().Secrets(remoteRemovalOptions.Namespace).List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return phase, err
	}
	if secretList != nil && secretList.Items != nil && len(secretList.Items) > 0 {
		phase = ResourceFound
		for _, resource := range secretList.Items {
			err := clientset.CoreV1().Secrets(remoteRemovalOptions.Namespace).Delete(resource.GetObjectMeta().GetName(), nil)
			if err != nil {
				phase = ResourceRemoveFailed
			} else {
				phase = ResourceRemoved
			}
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}

func deletePVC(remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceNotFound
	resourceList, err := clientset.CoreV1().PersistentVolumeClaims(remoteRemovalOptions.Namespace).List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return phase, err
	}
	if resourceList != nil && resourceList.Items != nil && len(resourceList.Items) > 0 {
		phase = ResourceFound
		for _, resource := range resourceList.Items {
			err := clientset.CoreV1().PersistentVolumeClaims(remoteRemovalOptions.Namespace).Delete(resource.GetObjectMeta().GetName(), nil)
			if err != nil {
				phase = ResourceRemoveFailed
			} else {
				phase = ResourceRemoved
			}
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}

func deleteServiceAccount(remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceNotFound
	resourceList, err := clientset.CoreV1().ServiceAccounts(remoteRemovalOptions.Namespace).List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return phase, err
	}
	if resourceList != nil && resourceList.Items != nil && len(resourceList.Items) > 0 {
		phase = ResourceFound
		for _, secret := range resourceList.Items {
			err := clientset.CoreV1().ServiceAccounts(remoteRemovalOptions.Namespace).Delete(secret.GetObjectMeta().GetName(), nil)
			if err != nil {
				phase = ResourceRemoveFailed
			} else {
				phase = ResourceRemoved
			}
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}

func deleteRoleBindings(remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceNotFound
	resourceList, err := clientset.RbacV1().RoleBindings(remoteRemovalOptions.Namespace).List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return phase, err
	}
	if resourceList != nil && resourceList.Items != nil && len(resourceList.Items) > 0 {
		phase = ResourceFound
		for _, resource := range resourceList.Items {
			err := clientset.RbacV1().RoleBindings(remoteRemovalOptions.Namespace).Delete(resource.GetObjectMeta().GetName(), nil)
			if err != nil {
				phase = ResourceRemoveFailed
			} else {
				phase = ResourceRemoved
			}
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}

func deleteTektonClusterRoleBindings(remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceNotFound
	resourceList, err := clientset.RbacV1().ClusterRoleBindings().List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return phase, err
	}
	if resourceList != nil && resourceList.Items != nil && len(resourceList.Items) > 0 {
		phase = ResourceFound
		for _, resource := range resourceList.Items {
			err := clientset.RbacV1().ClusterRoleBindings().Delete(resource.GetObjectMeta().GetName(), nil)
			if err != nil {
				phase = ResourceRemoveFailed
			} else {
				phase = ResourceRemoved
			}
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}

func deleteIngress(remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceNotFound
	resourceList, err := clientset.ExtensionsV1beta1().Ingresses(remoteRemovalOptions.Namespace).List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return phase, err
	}
	if resourceList != nil && resourceList.Items != nil && len(resourceList.Items) > 0 {
		phase = ResourceFound
		for _, secret := range resourceList.Items {
			err := clientset.ExtensionsV1beta1().Ingresses(remoteRemovalOptions.Namespace).Delete(secret.GetObjectMeta().GetName(), nil)
			if err != nil {
				phase = ResourceRemoveFailed
			} else {
				phase = ResourceRemoved
			}
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}

func deleteRoute(config *restclient.Config, remoteRemovalOptions *RemoveDeploymentOptions, clientset *kubernetes.Clientset, labelSelector string) (int, error) {
	phase := ResourceRemoveFailed
	routev1client, err := routev1.NewForConfig(config)
	if err != nil {
		return phase, err
	}
	resourceList, err := routev1client.Routes(remoteRemovalOptions.Namespace).List(
		v1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return ResourceNotFound, err
	}
	if resourceList != nil && resourceList.Items != nil && len(resourceList.Items) > 0 {
		phase = ResourceFound
		for _, secret := range resourceList.Items {
			err := routev1client.Routes(remoteRemovalOptions.Namespace).Delete(secret.GetObjectMeta().GetName(), nil)
			if err != nil {
				logr.Trace(secret.GetObjectMeta().GetName())
				phase = ResourceRemoveFailed
			} else {
				phase = ResourceRemoved
			}
		}
	} else {
		phase = ResourceNotFound
	}
	return phase, nil
}
