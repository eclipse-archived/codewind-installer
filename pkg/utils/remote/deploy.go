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
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/eclipse/codewind-installer/pkg/utils/remote/kube"
	logr "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// DeployOptions : Keycloak initial config
type DeployOptions struct {
	Namespace             string
	IngressDomain         string
	InstallKeycloak       bool
	KeycloakUser          string
	KeycloakPassword      string
	KeycloakDevUser       string
	KeycloakDevPassword   string
	KeycloakRealm         string
	KeycloakClient        string
	KeycloakSecure        bool
	KeycloakTLSSecure     bool
	GateKeeperTLSSecure   bool
	CodewindSessionSecret string
	ClientSecret          string
}

// DeploymentResult : Ingress root URLs
type DeploymentResult struct {
	GatekeeperURL string
	KeycloakURL   string
}

// DeployRemote : InstallRemote
func DeployRemote(remoteDeployOptions *DeployOptions) (*DeploymentResult, *RemInstError) {

	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
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

	// Get the ingress host
	ingressDomain := remoteDeployOptions.IngressDomain

	// Use a supplied ingress if one was not installed
	if remoteDeployOptions.IngressDomain == "" && !onOpenShift {
		logr.Infof("Attempting to discover Ingress Domain")
		svcList := clientset.CoreV1().Services("ingress-nginx")
		svc, err := svcList.List(v1.ListOptions{})
		if err == nil && svc != nil && svc.Items != nil {
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

	logr.Errorln("TODO : Build PVC, Acct and Secret")
	ownerReferenceUID = uuid.NewUUID()
	workspacePVC := "codewind"
	serviceAccountName := "codewind"
	secretName := "codewind"

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
		ServiceAccountName: serviceAccountName,
		PullSecret:         secretName,
		OwnerReferenceName: ownerReferenceName,
		OwnerReferenceUID:  ownerReferenceUID,
		Privileged:         true,
		Ingress:            "-" + workspaceID + "-" + ingressDomain,
		OnOpenShift:        onOpenShift,
	}

	err = DeployKeycloak(config, clientset, codewindInstance, remoteDeployOptions, onOpenShift)
	if err != nil {
		log.Printf("Codewind Keycloak failed, exiting...")
		os.Exit(1)
	}

	err = SetupKeycloak(codewindInstance, remoteDeployOptions)
	if err != nil {
		log.Printf("Codewind Keycloak configuration failed, exiting...")
		os.Exit(1)
	}

	err = DeployPFE(config, clientset, codewindInstance, remoteDeployOptions)
	if err != nil {
		log.Printf("Codewind deployment failed, exiting...")
		os.Exit(1)
	}

	err = DeployPerformance(clientset, codewindInstance, remoteDeployOptions)
	if err != nil {
		log.Printf("Codewind deployment failed, exiting...")
		os.Exit(1)
	}

	err = DeployGatekeeper(config, clientset, codewindInstance, remoteDeployOptions)
	if err != nil {
		log.Printf("Codewind Gatekeeper deployment failed, exiting...")
		os.Exit(1)
	}

	gatekeeperURL := GatekeeperPrefix + codewindInstance.Ingress
	keycloakURL := KeycloakPrefix + codewindInstance.Ingress

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
