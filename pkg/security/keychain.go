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
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/globals"
	"github.com/zalando/go-keyring"
)

var insecureKeyringDir = connections.GetConnectionConfigDir()

// KeyringSecret : Secret
type KeyringSecret struct {
	Service  []byte `json:"service"`
	Username []byte `json:"username"`
	Password []byte `json:"password"`
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

	keyringErr := StoreSecretInKeyring(conID, uName, pass)
	if keyringErr != nil {
		return keyringErr
	}

	/// check password can be retrieved
	secret, secErr := SecKeyGetSecret(conID, uName)
	if secErr != nil {
		return secErr
	}
	if secret != pass {
		secErr := errors.New("Saved password does not match retrieved password")
		return &SecError{errOpPasswordRead, secErr, secErr.Error()}
	}

	return nil
}

// SecKeyGetSecret : retrieve secret / credentials from the keyring
func SecKeyGetSecret(connectionID, username string) (string, *SecError) {
	conID := strings.TrimSpace(strings.ToLower(connectionID))
	uName := strings.TrimSpace(strings.ToLower(username))
	secret, err := GetSecretFromKeyring(conID, uName)
	if err != nil {
		return "", err
	}
	return secret, nil
}

// StoreSecretInKeyring stores the secret in either the system keyring or our insecure keyring.
func StoreSecretInKeyring(connectionID, uName, pass string) *SecError {
	service := connectionIDToService(connectionID)
	if globals.UseInsecureKeyring {
		_, statErr := os.Stat(GetPathToInsecureKeyring())
		if os.IsNotExist(statErr) {
			mkdirErr := os.MkdirAll(insecureKeyringDir, 0600)
			if mkdirErr != nil {
				return &SecError{errOpInsecureKeyring, mkdirErr, mkdirErr.Error()}
			}
			_, openFileErr := os.OpenFile(GetPathToInsecureKeyring(), os.O_CREATE, 0600)
			if openFileErr != nil {
				return &SecError{errOpInsecureKeyring, openFileErr, openFileErr.Error()}
			}
		}

		existingSecrets := []KeyringSecret{}
		file, readErr := ioutil.ReadFile(GetPathToInsecureKeyring())
		if readErr != nil {
			return &SecError{errOpInsecureKeyring, readErr, readErr.Error()}
		}
		if len(file) != 0 {
			unmarshalErr := json.Unmarshal([]byte(file), &existingSecrets)
			if unmarshalErr != nil {
				return &SecError{errOpInsecureKeyring, unmarshalErr, unmarshalErr.Error()}
			}
		}
		newSecret := KeyringSecret{
			Service:  []byte(service),
			Username: []byte(uName),
			Password: []byte(pass),
		}
		indexOfSecretToUpdate := -1

		for i, existingSecret := range existingSecrets {
			if doSecretsMatch(existingSecret, newSecret) {
				indexOfSecretToUpdate = i
			}
		}
		if indexOfSecretToUpdate > -1 {
			// remove existing secret
			existingSecrets = append(existingSecrets[:indexOfSecretToUpdate], existingSecrets[indexOfSecretToUpdate+1:]...)
		}
		secrets := append(existingSecrets, newSecret)
		body, marshallErr := json.MarshalIndent(secrets, "", "\t")
		if marshallErr != nil {
			return &SecError{errOpInsecureKeyring, marshallErr, marshallErr.Error()}
		}
		writeErr := ioutil.WriteFile(GetPathToInsecureKeyring(), body, 0644)
		if writeErr != nil {
			return &SecError{errOpInsecureKeyring, writeErr, writeErr.Error()}
		}
		return nil
	}

	// else store it in system keyring
	err := keyring.Set(service, uName, pass)
	if err != nil {
		return &SecError{errOpKeyring, err, err.Error()}
	}
	return nil
}

// GetSecretFromKeyring gets the secret from either the system keyring or our insecure keyring.
func GetSecretFromKeyring(connectionID, uName string) (string, *SecError) {
	service := connectionIDToService(connectionID)
	if globals.UseInsecureKeyring {
		secrets, readErr := readInsecureKeyring()
		if readErr != nil {
			return "", readErr
		}
		for _, secret := range secrets {
			sameService := string(secret.Service) == service
			sameUsername := string(secret.Username) == uName
			matchingSecret := sameService && sameUsername
			if matchingSecret {
				return string(secret.Password), nil
			}
		}
		err := errors.New(textSecretNotFound)
		return "", &SecError{errOpInsecureKeyring, err, err.Error()}
	}
	// else get from system keyring
	secret, err := keyring.Get(service, uName)
	if err != nil {
		if err == keyring.ErrNotFound {
			errNotFound := errors.New(textSecretNotFound)
			return "", &SecError{errOpKeyringSecretNotFound, errNotFound, errNotFound.Error()}
		}
		return "", &SecError{errOpKeyring, err, err.Error()}
	}
	return secret, nil
}

// DeleteSecretFromKeyring deletes the secret from either the system keyring or our insecure keyring.
func DeleteSecretFromKeyring(connectionID, uName string) *SecError {
	service := connectionIDToService(connectionID)
	if globals.UseInsecureKeyring {
		secrets, readErr := readInsecureKeyring()
		if readErr != nil {
			return readErr
		}
		indexOfSecretToDelete := -1
		for i, secret := range secrets {
			sameService := string(secret.Service) == service
			sameUsername := string(secret.Username) == uName
			matchingSecret := sameService && sameUsername
			if matchingSecret {
				indexOfSecretToDelete = i
			}
		}
		if indexOfSecretToDelete == -1 {
			err := errors.New(textSecretNotFound)
			return &SecError{errOpInsecureKeyring, err, err.Error()}
		}
		// remove existing secret
		secrets = append(secrets[:indexOfSecretToDelete], secrets[indexOfSecretToDelete+1:]...)
		if len(secrets) == 0 {
			err := os.Remove(GetPathToInsecureKeyring())
			if err != nil {
				return &SecError{errOpInsecureKeyring, err, err.Error()}
			}
			return nil
		}
		body, marshallErr := json.MarshalIndent(secrets, "", "\t")
		if marshallErr != nil {
			return &SecError{errOpInsecureKeyring, marshallErr, marshallErr.Error()}
		}
		writeErr := ioutil.WriteFile(GetPathToInsecureKeyring(), body, 0644)
		if writeErr != nil {
			return &SecError{errOpInsecureKeyring, writeErr, writeErr.Error()}
		}
		return nil
	}
	// else delete from system keyring
	err := keyring.Delete(service, uName)
	if err != nil {
		if err == keyring.ErrNotFound {
			errNotFound := errors.New(textSecretNotFound)
			return &SecError{errOpKeyringSecretNotFound, errNotFound, errNotFound.Error()}
		}
		return &SecError{errOpKeyring, err, err.Error()}
	}
	return nil
}

func readInsecureKeyring() ([]KeyringSecret, *SecError) {
	file, readErr := ioutil.ReadFile(GetPathToInsecureKeyring())
	if readErr != nil {
		if os.IsNotExist(readErr) {
			err := errors.New(textSecretNotFound)
			return nil, &SecError{errOpInsecureKeyring, err, err.Error()}
		}
		return nil, &SecError{errOpInsecureKeyring, readErr, readErr.Error()}
	}
	secrets := []KeyringSecret{}
	if len(file) != 0 {
		unmarshalErr := json.Unmarshal([]byte(file), &secrets)
		if unmarshalErr != nil {
			return nil, &SecError{errOpInsecureKeyring, unmarshalErr, unmarshalErr.Error()}
		}
	}
	return secrets, nil
}

func connectionIDToService(connectionID string) string {
	conID := strings.TrimSpace(strings.ToLower(connectionID))
	return KeyringServiceName + "." + conID
}

// GetPathToInsecureKeyring gets the path to the insecureKeychain.json
func GetPathToInsecureKeyring() string {
	return path.Join(insecureKeyringDir, "insecureKeychain.json")
}

func doSecretsMatch(secret1, secret2 KeyringSecret) bool {
	sameService := bytes.Compare(secret1.Service, secret2.Service) == 0
	sameUsername := bytes.Compare(secret1.Username, secret2.Username) == 0
	matchingSecret := sameService && sameUsername
	return matchingSecret
}
