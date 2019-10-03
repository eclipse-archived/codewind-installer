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

	"github.com/urfave/cli"
	"github.com/zalando/go-keyring"
)

// KeyringSecret : Secret
type KeyringSecret struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
}

// SecKeyUpdate : Creates or updates a key in the platforms keyring
func SecKeyUpdate(c *cli.Context) *SecError {
	depid := strings.TrimSpace(strings.ToLower(c.String("depid")))
	username := strings.TrimSpace(c.String("username"))
	password := strings.TrimSpace(c.String("password"))
	err := keyring.Set(KeyringServiceName+"."+depid, username, password)
	if err != nil {
		return &SecError{errOpKeyring, err, err.Error()}
	}
	return nil
}

// SecKeyGetSecret : retrieve secret / credentials from the keyring
func SecKeyGetSecret(c *cli.Context) (string, *SecError) {
	depid := strings.TrimSpace(c.String("depid"))
	username := strings.TrimSpace(c.String("username"))
	secret, err := keyring.Get(KeyringServiceName+"."+depid, username)
	if err != nil {
		return "", &SecError{errOpKeyring, err, err.Error()}
	}
	return secret, nil
}
