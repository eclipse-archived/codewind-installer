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
	"os"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/globals"

	"github.com/stretchr/testify/assert"
)

const testPassword = "pAss%-w0rd-&'cha*s"
const testPasswordUpdated = "pAss%-w0rd-&'cha*s-with_more_chars"

func Test_Keychain_Secure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	var originalUseInsecureKeyring = globals.UseInsecureKeyring
	globals.SetUseInsecureKeyring(false)

	// remove test key if it already exists
	DeleteSecretFromKeyring(testConnection, testUsername)

	t.Run("Secret cannot be retrieved for an unknown account", func(t *testing.T) {
		retrievedSecret, err := SecKeyGetSecret(testConnection, testUsername)
		assert.Equal(t, err.Op, "sec_keyring_secret_not_found")
		assert.NotNil(t, err)
		assert.Equal(t, "", retrievedSecret)
	})

	t.Run("A new key can be created in the platform keychain", func(t *testing.T) {
		err := SecKeyUpdate(testConnection, testUsername, testPassword)
		assert.Nil(t, err)
	})

	t.Run("The secret can be retrieved from the keychain", func(t *testing.T) {
		storedSecret, err := SecKeyGetSecret(testConnection, testUsername)
		assert.Nil(t, err)
		assert.Equal(t, testPassword, storedSecret)
	})

	t.Run("An existing key in the keychain can be updated", func(t *testing.T) {
		err := SecKeyUpdate(testConnection, testUsername, testPasswordUpdated)
		assert.Nil(t, err)
	})

	t.Run("Retrieved secret matches the saved secret", func(t *testing.T) {
		storedSecret, err := SecKeyGetSecret(testConnection, testUsername)
		assert.Nil(t, err)
		assert.Equal(t, testPasswordUpdated, storedSecret)
	})

	t.Run("Test keyring entry can be removed", func(t *testing.T) {
		err := DeleteSecretFromKeyring(testConnection, testUsername)
		assert.Nil(t, err)
	})

	t.Run("Test keyring returns an error when trying to delete a non-existent secret", func(t *testing.T) {
		err := DeleteSecretFromKeyring(testConnection, testUsername)
		assert.NotNil(t, err)
		assert.Equal(t, err.Op, "sec_keyring_secret_not_found")
		assert.Equal(t, "Secret not found in keyring", err.Desc)
	})

	globals.SetUseInsecureKeyring(originalUseInsecureKeyring)
}

func Test_Keychain_Insecure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	var originalUseInsecureKeyring = globals.UseInsecureKeyring
	globals.SetUseInsecureKeyring(true)

	// remove insecureKeychain.json if it already exists
	os.Remove(GetPathToInsecureKeyring())

	t.Run("Secret cannot be retrieved when keychain file does not exist", func(t *testing.T) {
		retrievedSecret, err := SecKeyGetSecret(testConnection, testUsername)
		assert.NotNil(t, err)
		assert.Contains(t, err.Desc, "Secret not found in keyring")
		assert.Equal(t, "", retrievedSecret)
	})

	t.Run("A new secret is stored in a new keychain file when the keychain file didn't exist", func(t *testing.T) {
		err := SecKeyUpdate(testConnection, testUsername, testPassword)
		assert.Nil(t, err)
		assert.FileExists(t, GetPathToInsecureKeyring())
	})

	t.Run("The secret can be retrieved from the keychain", func(t *testing.T) {
		storedSecret, err := SecKeyGetSecret(testConnection, testUsername)
		assert.Nil(t, err)
		assert.Equal(t, testPassword, storedSecret)
	})

	t.Run("Secret cannot be retrieved for an unknown account", func(t *testing.T) {
		retrievedSecret, err := SecKeyGetSecret(testConnection, "unknownaccount")
		assert.NotNil(t, err)
		assert.Equal(t, "Secret not found in keyring", err.Desc)
		assert.Equal(t, "", retrievedSecret)
	})

	t.Run("An existing key in the keychain can be updated", func(t *testing.T) {
		err := SecKeyUpdate(testConnection, testUsername, testPasswordUpdated)
		assert.Nil(t, err)
	})

	t.Run("Retrieved secret matches the saved secret", func(t *testing.T) {
		storedSecret, err := SecKeyGetSecret(testConnection, testUsername)
		assert.Nil(t, err)
		assert.Equal(t, testPasswordUpdated, storedSecret)
	})

	t.Run("Test keyring returns an error when trying to delete from a non-existent keyring", func(t *testing.T) {
		err := DeleteSecretFromKeyring("mockConnectionID", "mockUsername")
		assert.NotNil(t, err)
		assert.Equal(t, "Secret not found in keyring", err.Desc)
	})

	t.Run("Secret can be removed without deleting keychain", func(t *testing.T) {
		StoreSecretInKeyring("mockConnectionID", "mockUsername", "mockPassword")

		err := DeleteSecretFromKeyring("mockConnectionID", "mockUsername")
		assert.Nil(t, err)

		// check this secret has been deleted
		secretFromThisTest, err := GetSecretFromKeyring("mockConnectionID", "mockUsername")
		assert.Equal(t, "", secretFromThisTest)
		assert.Equal(t, "Secret not found in keyring", err.Desc)

		// check the secret created before this test has not been deleted
		existingSecret, err2 := GetSecretFromKeyring(testConnection, testUsername)
		assert.Equal(t, testPasswordUpdated, existingSecret)
		assert.Nil(t, err2)
		assert.FileExists(t, GetPathToInsecureKeyring())
	})

	t.Run("Keychain is deleted when last secret is removed", func(t *testing.T) {
		err := DeleteSecretFromKeyring(testConnection, testUsername)
		assert.Nil(t, err)
		assert.True(t, noFileExists(GetPathToInsecureKeyring()))
	})

	t.Run("Test keyring returns an error when trying to delete from a non-existent keyring", func(t *testing.T) {
		err := DeleteSecretFromKeyring(testConnection, testUsername)
		assert.NotNil(t, err)
		assert.Contains(t, err.Desc, "Secret not found in keyring")
	})

	// remove insecureKeychain.json if it still exists
	os.Remove(GetPathToInsecureKeyring())

	globals.SetUseInsecureKeyring(originalUseInsecureKeyring)
}

func noFileExists(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		return true
	}
	if info.IsDir() {
		return true
	}
	return false
}
