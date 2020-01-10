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

	"github.com/eclipse/codewind-installer/pkg/connections"
	cwerrors "github.com/eclipse/codewind-installer/pkg/errors"
	"github.com/zalando/go-keyring"
)

// KeyringSecret : Secret
type KeyringSecret struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
}

// SecKeyUpdate : Creates or updates a key in the platforms keyring
func SecKeyUpdate(connectionID string, username string, password string) *cwerrors.BasicError {

	conID := strings.TrimSpace(strings.ToLower(connectionID))
	uName := strings.TrimSpace(strings.ToLower(username))
	pass := strings.TrimSpace(password)

	// check connection has been registered
	_, conErr := connections.GetConnectionByID(conID)
	if conErr != nil {
		err := errors.New("Connection " + strings.ToUpper(conID) + " not found")
		return &cwerrors.BasicError{errOpNotFound, err, conErr.Error()}
	}

	err := keyring.Set(KeyringServiceName+"."+conID, uName, pass)
	if err != nil {
		return &cwerrors.BasicError{errOpKeyring, err, err.Error()}
	}

	/// check password can be retrieved
	secret, secErr := SecKeyGetSecret(conID, uName)
	if err != nil {
		return secErr
	}
	if secret != pass {
		secErr := errors.New("Saved password does not match retrieved password")
		return &cwerrors.BasicError{errOpPasswordRead, secErr, secErr.Error()}
	}

	return nil
}

// SecKeyGetSecret : retrieve secret / credentials from the keyring
func SecKeyGetSecret(connectionID string, username string) (string, *cwerrors.BasicError) {

	conID := strings.TrimSpace(strings.ToLower(connectionID))
	uName := strings.TrimSpace(strings.ToLower(username))

	secret, err := keyring.Get(KeyringServiceName+"."+conID, uName)
	if err != nil {
		return "", &cwerrors.BasicError{errOpKeyring, err, err.Error()}
	}
	return secret, nil
}
