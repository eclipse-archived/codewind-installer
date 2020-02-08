/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

package remote

// MockCodewind is a standard configuration for a remote Codewind
var MockCodewind = Codewind{
	PFEName:            "pfe",
	PerformanceName:    "performance",
	GatekeeperName:     "gatekeeper",
	KeycloakName:       "keycloak",
	PFEImage:           "/pfe-image",
	PerformanceImage:   "performance-image",
	GatekeeperImage:    "gatekeeper-image",
	KeycloakImage:      "keycloak-image",
	Namespace:          "testspace",
	WorkspaceID:        "test-id",
	PVCName:            "test-pvc",
	ServiceAccountName: "test-sa",
	ServiceAccountKC:   "test-sakc",
	OwnerReferenceName: "test-refname",
	Privileged:         false,
	Ingress:            "ingress",
	RequestedIngress:   "test-requested-ingress",
	OnOpenShift:        false,
}
