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
	"errors"
	"strings"

	"github.com/eclipse/codewind-installer/utils/connections"
	"github.com/zalando/go-keyring"
)

// KeyringSecret : Secret
type KeyringSecret struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
}

// SecKeyUpdate : Creates or updates a key in the platforms keyring
func SecKeyUpdate(connectionID string, username string, password string) *SecError {

	conID := strings.TrimSpace(strings.ToLower(connectionID))
	uName := strings.TrimSpace(strings.ToLower(username))
	pass := strings.TrimSpace(password)

	// check connection has been registered
	_, conErr := connections.GetConnectionByID(conID)
	if conErr != nil {
		err := errors.New("Connection " + strings.ToUpper(conID) + " not found")
		return &SecError{errOpNotFound, err, conErr.Error()}
	}

	err := keyring.Set(KeyringServiceName+"."+conID, uName, pass)
	if err != nil {
		return &SecError{errOpKeyring, err, err.Error()}
	}
	return nil
}

// SecKeyGetSecret : retrieve secret / credentials from the keyring
func SecKeyGetSecret(connectionID string, username string) (string, *SecError) {

	conID := strings.TrimSpace(strings.ToLower(connectionID))
	uName := strings.TrimSpace(strings.ToLower(username))

	secret, err := keyring.Get(KeyringServiceName+"."+conID, uName)
	if err != nil {
		return "", &SecError{errOpKeyring, err, err.Error()}
	}
	return secret, nil
}
