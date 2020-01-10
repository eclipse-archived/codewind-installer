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

const (
	errOpNotFound        = "rem_not_found"
	errOpNoIngress       = "rem_no_ingress"
	errOpCreateNamespace = "rem_create_namespace"
)

const (
	errTargetNotFound   = "Target deployment not found"
	errNoIngressService = "Please check you have installed ingress-nginx into your Kubernetes environment or use the --ingress flag to set the domain"
)

// Result : status message
type Result struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
}
