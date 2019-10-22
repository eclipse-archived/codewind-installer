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

package security

import (
	"strings"

	"github.com/zalando/go-keyring"
)

// KeyringSecret : Secret
type KeyringSecret struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
}

// SecKeyUpdate : Creates or updates a key in the platforms keyring
func SecKeyUpdate(deploymentID string, username string, password string) *SecError {

	depID := strings.TrimSpace(strings.ToLower(deploymentID))
	uName := strings.TrimSpace(strings.ToLower(username))
	pass := strings.TrimSpace(password)

	err := keyring.Set(KeyringServiceName+"."+depID, uName, pass)
	if err != nil {
		return &SecError{errOpKeyring, err, err.Error()}
	}
	return nil
}

// SecKeyGetSecret : retrieve secret / credentials from the keyring
func SecKeyGetSecret(deploymentID string, username string) (string, *SecError) {

	depID := strings.TrimSpace(strings.ToLower(deploymentID))
	uName := strings.TrimSpace(strings.ToLower(username))

	secret, err := keyring.Get(KeyringServiceName+"."+depID, uName)
	if err != nil {
		return "", &SecError{errOpKeyring, err, err.Error()}
	}
	return secret, nil
}
