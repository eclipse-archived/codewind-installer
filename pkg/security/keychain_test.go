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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

const testPassword = "pAss%-w0rd-&'cha*s"
const testPasswordUpdated = "pAss%-w0rd-&'cha*s-with_more_chars"

func Test_Keychain(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	// remove test key if one exists
	keyring.Delete(strings.ToLower(KeyringServiceName+"."+testConnection), testUsername)

	t.Run("Secret can not be retrieved for an unknown account", func(t *testing.T) {
		retrievedSecret, err := SecKeyGetSecret(testConnection, testUsername)
		if err == nil {
			t.Fail()
		}
		assert.Equal(t, "", retrievedSecret)
	})

	t.Run("A new key can be created in the platform keychain", func(t *testing.T) {
		err := SecKeyUpdate(testConnection, testUsername, testPassword)
		if err != nil {
			t.Fail()
		}
	})

	t.Run("The secret can be retrieved from the keychain", func(t *testing.T) {
		storedSecret, err := SecKeyGetSecret(testConnection, testUsername)
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, testPassword, storedSecret)
	})

	t.Run("An existing key in the keychain can be updated", func(t *testing.T) {
		err := SecKeyUpdate(testConnection, testUsername, testPasswordUpdated)
		if err != nil {
			t.Fail()
		}
	})

	t.Run("Retrieved secret matches the saved secret", func(t *testing.T) {
		storedSecret, err := SecKeyGetSecret(testConnection, testUsername)
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, testPasswordUpdated, storedSecret)
	})

	t.Run("Test keyring entry can be removed", func(t *testing.T) {
		err := keyring.Delete(strings.ToLower(KeyringServiceName+"."+testConnection), testUsername)
		if err != nil {
			t.Fail()
		}
	})

}
