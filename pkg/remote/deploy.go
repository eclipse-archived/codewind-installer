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
	"os"
	"strconv"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/remote/kube"
	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Required for Kube clusters which use auth plugins
)

// DeployOptions : Keycloak initial config
type DeployOptions struct {
	Namespace             string
	IngressDomain         string
	KeycloakUser          string
	KeycloakPassword      string
	KeycloakDevUser       string
	KeycloakDevPassword   string
	KeycloakRealm         string
	KeycloakClient        string
	KeycloakSecure        bool
	KeycloakTLSSecure     bool
	KeycloakURL           string
	KeycloakHost          string
	KeycloakOnly          bool
	GateKeeperTLSSecure   bool
	CodewindSessionSecret string
	ClientSecret          string
	CodewindPVCSize       string
	LogLevel              string
}

// DeploymentResult : Ingress root URLs
type DeploymentResult struct {
	GatekeeperURL string
	KeycloakURL   string
}

// DeployRemote : InstallRemote
func DeployRemote(remoteDeployOptions *DeployOptions) (*DeploymentResult, *RemInstError) {
	config, err := getKubeConfig()
	if err != nil {
		logr.Infof("Unable to retrieve Kubernetes Config %v\n", err)
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logr.Infof("Unable to retrieve Kubernetes clientset %v\n", err)
		return nil, &RemInstError{errOpNotFound, err, err.Error()}
	}

	namespace := remoteDeployOptions.Namespace
	// Get the current namespace
	if namespace == "" {
		namespace = kube.GetCurrentNamespace()
	}

	// Check if namespace exists
	logr.Infof("Checking namespace %v exists\n", namespace)
	_, err = clientset.CoreV1().Namespaces().Get(namespace, v1.GetOptions{})
	if err != nil {
		logr.Infof("Creating %v namespace\n", namespace)
		// create the namespace
		deploymentNamespace := corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}

		// insert the namespace
		requestedNamespace, err := clientset.CoreV1().Namespaces().Create(&deploymentNamespace)
		if err != nil || requestedNamespace == nil {
			logr.Errorf("Unable to create %v namespace: %v", namespace, err)
			return nil, &RemInstError{errOpCreateNamespace, err, err.Error()}
		}
	}

	logr.Infof("Using namespace : %v\n", namespace)
	pfeImage, performanceImage, keycloakImage, gatekeeperImage := GetImages()

	logr.Infoln("Container images : ")
	logr.Infoln(pfeImage)
	logr.Infoln(performanceImage)
	logr.Infoln(keycloakImage)
	logr.Infoln(gatekeeperImage)

	// Determine if we're running on OpenShift or not.
	onOpenShift := kube.DetectOpenShift(config)
	logr.Infof("Running on openshift: %t\n", onOpenShift)

	workspaceID := strings.ToLower(strconv.FormatInt(utils.CreateTimestamp(), 36))

	// append workspaceID to the client name
	remoteDeployOptions.KeycloakClient = remoteDeployOptions.KeycloakClient + "-" + workspaceID

	// Get the ingress host
	ingressDomain := remoteDeployOptions.IngressDomain

	// Use a supplied ingress if one was not installed
	if remoteDeployOptions.IngressDomain == "" && !onOpenShift {
		logr.Infof("Attempting to discover Ingress Domain")
		svcList := clientset.CoreV1().Services("ingress-nginx")
		svc, err := svcList.List(v1.ListOptions{})
		if err == nil && svc != nil && svc.Items != nil && len(svc.Items) > 0 {
			ingressDomain = svc.Items[0].Spec.ClusterIP + ".nip.io"
		}
	}

	// Check ingress service installed
	if ingressDomain == "" {
		remoteInstError := errors.New(errNoIngressService)
		return nil, &RemInstError{errOpNoIngress, remoteInstError, remoteInstError.Error()}
	}

	logr.Infof("Using ingress domain: %v\n", ingressDomain)

	var ownerReferenceName string
	var ownerReferenceUID types.UID
	ownerReferenceName = "codewind" + workspaceID
	ownerReferenceUID = uuid.NewUUID()

	workspacePVC := PFEPrefix + "-pvc-" + workspaceID

	// Create the Codewind deployment object
	codewindInstance := Codewind{
		PFEName:            PFEPrefix + workspaceID,
		PFEImage:           pfeImage,
		PerformanceName:    PerformancePrefix + workspaceID,
		PerformanceImage:   performanceImage,
		KeycloakName:       KeycloakPrefix + workspaceID,
		KeycloakImage:      keycloakImage,
		GatekeeperName:     GatekeeperPrefix + workspaceID,
		GatekeeperImage:    gatekeeperImage,
		Namespace:          namespace,
		WorkspaceID:        workspaceID,
		PVCName:            workspacePVC,
		ServiceAccountName: "codewind-" + workspaceID, //  codewind-k39vwfk0
		ServiceAccountKC:   "keycloak-" + workspaceID, //  keycloak-k39vwfk0
		OwnerReferenceName: ownerReferenceName,
		OwnerReferenceUID:  ownerReferenceUID,
		Privileged:         true,
		Ingress:            "-" + workspaceID + "." + ingressDomain,
		RequestedIngress:   ingressDomain,
		OnOpenShift:        onOpenShift,
	}

	gatekeeperURL := GatekeeperPrefix + codewindInstance.Ingress
	keycloakURL := KeycloakPrefix + codewindInstance.Ingress

	// Create the Codewind service account
	if !remoteDeployOptions.KeycloakOnly {
		codewindServiceTemplate := CreateCodewindServiceAcct(codewindInstance, remoteDeployOptions)
		_, err = clientset.CoreV1().ServiceAccounts(namespace).Create(&codewindServiceTemplate)
		if err != nil {
			logr.Errorln("Creating service account failed, exiting...")
			logr.Errorln(err)
			os.Exit(1)
		}
	}

	// If we are not using an existing Keycloak, deploy one now
	if remoteDeployOptions.KeycloakURL == "" {
		keycloakServiceAccountTemplate := CreateKeycloakServiceAcct(codewindInstance, remoteDeployOptions)
		_, err = clientset.CoreV1().ServiceAccounts(namespace).Create(&keycloakServiceAccountTemplate)
		if err != nil {
			logr.Errorln("Creating Keycloak service account failed, exiting...")
			logr.Errorln(err)
			os.Exit(1)
		}
		err = DeployKeycloak(config, clientset, codewindInstance, remoteDeployOptions, onOpenShift)
		if err != nil {
			logr.Errorln("Codewind Keycloak failed, exiting...")
			os.Exit(1)
		}
		podSearch := "codewindWorkspace=" + codewindInstance.WorkspaceID + ",app=" + KeycloakPrefix
		ready := false
		for !ready {
			ready = WaitForPodReady(clientset, codewindInstance, podSearch, KeycloakPrefix+"-"+codewindInstance.WorkspaceID)
		}
	}

	err = SetupKeycloak(codewindInstance, remoteDeployOptions)
	if err != nil {
		logr.Errorln("Codewind Keycloak configuration failed, exiting...")
		os.Exit(1)
	}

	if remoteDeployOptions.KeycloakOnly {
		deploymentResult := DeploymentResult{
			KeycloakURL: keycloakURL,
		}
		return &deploymentResult, nil
	}

	err = DeployPFE(config, clientset, codewindInstance, remoteDeployOptions)
	if err != nil {
		logr.Errorln("Codewind deployment failed, exiting...")
		os.Exit(1)
	}

	podSearch := "codewindWorkspace=" + codewindInstance.WorkspaceID + ",app=" + PFEPrefix
	ready := false
	for !ready {
		ready = WaitForPodReady(clientset, codewindInstance, podSearch, PFEPrefix+"-"+codewindInstance.WorkspaceID)
	}

	err = DeployPerformance(clientset, codewindInstance, remoteDeployOptions)
	if err != nil {
		logr.Errorln("Codewind deployment failed, exiting...")
		os.Exit(1)
	}

	podSearch = "codewindWorkspace=" + codewindInstance.WorkspaceID + ",app=" + PerformancePrefix
	ready = false
	for !ready {
		ready = WaitForPodReady(clientset, codewindInstance, podSearch, PerformancePrefix+"-"+codewindInstance.WorkspaceID)
	}

	err = DeployGatekeeper(config, clientset, codewindInstance, remoteDeployOptions)
	if err != nil {
		logr.Errorln("Codewind Gatekeeper deployment failed, exiting...")
		os.Exit(1)
	}

	podSearch = "codewindWorkspace=" + codewindInstance.WorkspaceID + ",app=" + GatekeeperPrefix
	ready = false
	for !ready {
		ready = WaitForPodReady(clientset, codewindInstance, podSearch, GatekeeperPrefix+"-"+codewindInstance.WorkspaceID)
	}

	if remoteDeployOptions.GateKeeperTLSSecure {
		gatekeeperURL = "https://" + gatekeeperURL
	} else {
		gatekeeperURL = "http://" + gatekeeperURL
	}

	if remoteDeployOptions.KeycloakTLSSecure {
		keycloakURL = "https://" + keycloakURL
	} else {
		keycloakURL = "http://" + keycloakURL
	}

	deploymentResult := DeploymentResult{
		GatekeeperURL: gatekeeperURL,
		KeycloakURL:   keycloakURL,
	}

	return &deploymentResult, nil
}
