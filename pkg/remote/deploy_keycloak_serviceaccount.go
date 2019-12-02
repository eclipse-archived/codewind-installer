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
	logr "github.com/sirupsen/logrus"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateKeycloakServiceAcct : Create service account
func CreateKeycloakServiceAcct(codewind Codewind, deployOptions *DeployOptions) coreV1.ServiceAccount {
	logr.Infof("Creating service account definition '%v'", codewind.ServiceAccountKC)

	labels := map[string]string{
		"codewindWorkspace": codewind.WorkspaceID,
		"app":               codewind.ServiceAccountKC,
	}
	svc := coreV1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   codewind.ServiceAccountKC,
			Labels: labels,
		},
		Secrets: nil,
	}
	return svc
}
