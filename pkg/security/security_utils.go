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
	"encoding/json"
)

// KeycloakMasterRealm : master realm name
const KeycloakMasterRealm string = "master"

// KeycloakAdminClientID : master realm name
const KeycloakAdminClientID string = "admin-cli"

// CodewindCliID : master realm name
const CodewindCliID string = "codewind-cli"

// CodewindClientID : master realm name
const CodewindClientID string = "codewind-backend"

// KeyringServiceName : name
const KeyringServiceName string = "org.eclipse.codewind"

// SecError : Security package errors
type SecError struct {
	Op   string
	Err  error
	Desc string
}

const (
	errOpConnection            = "sec_connection"               // Connection failed
	errOpResponse              = "sec_response"                 // Bad response
	errOpResponseFormat        = "sec_bodyparser"               // Parse errors
	errOpNotFound              = "sec_notfound"                 // No matching search results
	errOpCreate                = "sec_create"                   // Create failed
	errOpPassword              = "sec_passwordcontent"          // Password formatting
	errOpHostname              = "sec_badhostname"              // Bad hostname / url
	errOpKeyring               = "sec_keyring"                  // Keyring operations
	errOpKeyringSecretNotFound = "sec_keyring_secret_not_found" // ErrNotFound received from a Keyring operation
	errOpInsecureKeyring       = "sec_insecure_keyring"         // Insecure Keyring operations
	errOpConConfig             = "sec_con_config"               // Connection configuration errors
	errOpCLICommand            = "sec_cli_options"              // Invalid command line options
	errOpPasswordRead          = "sec_password_read"            // Unable to fetch password
)

const (
	textBadPassword    = "Passwords must not contains quoted characters"
	textUserNotFound   = "Registered User not found"
	textUnableToParse  = "Unable to parse Keycloak response"
	textInvalidOptions = "Invalid or missing command line options"
	textAuthIsDown     = "Authentication service unavailable"
	textSecretNotFound = "Secret not found in keyring"
)

// SecError : Error formatted in JSON containing an errorOp and a description from
// either a fault condition in the CLI, or an error payload from a REST request
func (se *SecError) Error() string {
	type Output struct {
		Operation   string `json:"error"`
		Description string `json:"error_description"`
	}
	tempOutput := &Output{Operation: se.Op, Description: se.Err.Error()}
	jsonError, _ := json.Marshal(tempOutput)
	return string(jsonError)
}

// KeycloakAPIError : Error responses from Keycloak
type KeycloakAPIError struct {
	HTTPStatus       int
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorMessage     string `json:"errorMessage"`
}

// Result : status messaqe
type Result struct {
	Status string `json:"status"`
}

// parseKeycloakError : parse the JSON response from Keycloak
func parseKeycloakError(body string, httpCode int) *KeycloakAPIError {
	keycloakAPIError := KeycloakAPIError{}
	keycloakAPIError.HTTPStatus = httpCode
	err := json.Unmarshal([]byte(body), &keycloakAPIError)
	if err != nil {
		keycloakAPIError.ErrorDescription = "Bad JSON in error response from keycloak"
	}
	// copy message into description if one is not set
	if keycloakAPIError.ErrorMessage != "" && keycloakAPIError.ErrorDescription == "" {
		keycloakAPIError.ErrorDescription = keycloakAPIError.ErrorMessage
	}
	return &keycloakAPIError
}

// IsSecretNotFoundError : Test whether a secret error is due to the secret not exisiting.
func IsSecretNotFoundError(se *SecError) bool {
	return se.Desc == textSecretNotFound
}
