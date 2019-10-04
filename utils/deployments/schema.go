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

// File schema versions of the Deployments configuration file

package deployments

// DeploymentConfigV0 : DeploymentsConfig Schema Version 0
type DeploymentConfigV0 struct {
	Active      string         `json:"active"`
	Deployments []DeploymentV0 `json:"deployments"`
}

// DeploymentConfigV1 : Deployments Schema Version 1
type DeploymentConfigV1 struct {
	SchemaVersion int            `json:"schemaversion"`
	Active        string         `json:"active"`
	Deployments   []DeploymentV1 `json:"deployments"`
}

// DeploymentV0 : Deployments Schema Version 0
type DeploymentV0 struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	AuthURL  string `json:"auth"`
	Realm    string `json:"realm"`
	ClientID string `json:"client_id"`
}

// DeploymentV1 : Deployments Schema Version 1
type DeploymentV1 struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	AuthURL  string `json:"auth"`
	Realm    string `json:"realm"`
	ClientID string `json:"client_id"`
}
