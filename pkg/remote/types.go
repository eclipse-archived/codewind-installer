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

import "k8s.io/apimachinery/pkg/types"

// Codewind represents a Codewind instance: name, namespace, volume, serviceaccount, and pull secrets
type Codewind struct {
	PFEName            string
	PerformanceName    string
	GatekeeperName     string
	KeycloakName       string
	PFEImage           string
	PerformanceImage   string
	GatekeeperImage    string
	KeycloakImage      string
	Namespace          string
	WorkspaceID        string
	PVCName            string
	ServiceAccountName string
	ServiceAccountKC   string
	OwnerReferenceName string
	OwnerReferenceUID  types.UID
	Privileged         bool
	Ingress            string
	RequestedIngress   string // resolved where possible or set by cli flag
	OnOpenShift        bool
}

// ServiceAccountPatch contains an array of imagePullSecrets that will be patched into a Kubernetes service account
type ServiceAccountPatch struct {
	ImagePullSecrets *[]ImagePullSecret `json:"imagePullSecrets"`
}

// ImagePullSecret represents a Kubernetes imagePullSecret
type ImagePullSecret struct {
	Name string `json:"name"`
}
