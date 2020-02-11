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
	"github.com/eclipse/codewind-installer/pkg/appconstants"
	corev1 "k8s.io/api/core/v1"
)

const (
	// PFEPrefix is the prefix all PFE-related resources: deployment, service, and ingress/route
	PFEPrefix = "codewind-pfe"

	// PerformancePrefix is the prefix for all performance-dashboard related resources: deployment and service
	PerformancePrefix = "codewind-performance"

	// KeycloakPrefix is the prefix for all keycloak related resources: deployment and service
	KeycloakPrefix = "codewind-keycloak"

	// GatekeeperPrefix is the prefix for all gatekeeper related resources: deployment and service
	GatekeeperPrefix = "codewind-gatekeeper"

	// PFEImage is the docker image that will be used in the Codewind-PFE pod
	PFEImage = "eclipse/codewind-pfe-amd64"

	// PerformanceImage is the docker image that will be used in the Performance dashboard pod
	PerformanceImage = "eclipse/codewind-performance-amd64"

	// KeycloakImage is the docker image that will be used in the Codewind-Keycloak pod
	KeycloakImage = "eclipse/codewind-keycloak-amd64"

	// GatekeeperImage is the docker image that will be used in the Codewind-Gatekeeper pod
	GatekeeperImage = "eclipse/codewind-gatekeeper-amd64"

	// PFEImageTag is the image tag associated with the docker image that's used for Codewind-PFE
	PFEImageTag = "latest"

	// PerformanceTag is the image tag associated with the docker image that's used for the Performance dashboard
	PerformanceTag = "latest"

	// KeycloakImageTag is the image tag associated with the docker image that's used for Codewind-Keycloak
	KeycloakImageTag = "latest"

	// GatekeeperImageTag is the image tag associated with the docker image that's used for Codewind-Gatekeeper
	GatekeeperImageTag = "latest"

	// ImagePullPolicy is the pull policy used for all containers in Codewind, defaults to Always
	ImagePullPolicy = corev1.PullAlways

	// PFEContainerPort is the port at which Codewind-PFE is exposed
	PFEContainerPort = 9191

	// PerformanceContainerPort is the port at which the Performance dashboard is exposed
	PerformanceContainerPort = 9095

	// KeycloakContainerPort is the port at which Keycloak is exposed
	KeycloakContainerPort = 8080

	// GatekeeperContainerPort is the port at which the Gatekeeper is exposed
	GatekeeperContainerPort = 9096

	// CodewindRolesName will include the workspaceID when deployed
	CodewindRolesName = "eclipse-codewind-" + appconstants.VersionNum

	// CodewindRoleBindingNamePrefix will include the workspaceID when deployed
	CodewindRoleBindingNamePrefix = "codewind-rolebinding"

	// CodewindTektonClusterRoleBindingName : Tekton, cluster role binding
	CodewindTektonClusterRoleBindingName = "codewind-tekton-rolebinding"

	// CodewindTektonClusterRolesName : Tekton, cluster role
	CodewindTektonClusterRolesName = "codewind-tekton"

	// ROKSStorageClass references the storage class to use on ROKS (OpenShift on IKS)
	ROKSStorageClass = "ibmc-file-bronze"
)
