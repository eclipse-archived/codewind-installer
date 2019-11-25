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

package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

// SecurityTokenGet : Authenticate and retrieve an access_token
func SecurityTokenGet(c *cli.Context) {
	auth, err := security.SecAuthenticate(http.DefaultClient, c, "", "")
	if err == nil && auth != nil {
		utils.PrettyPrintJSON(auth)
	} else {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

// SecurityTokenRefresh : Refresh the access token the cached refresh token
func SecurityTokenRefresh(c *cli.Context) {
	authTokens, secErr := security.SecRefreshTokens(http.DefaultClient, c)
	if secErr == nil && authTokens != nil {
		utils.PrettyPrintJSON(authTokens)
	} else {
		fmt.Println(secErr.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

// SecurityCreateRealm : Create a realm in Keycloak
func SecurityCreateRealm(c *cli.Context) {
	err := security.SecRealmCreate(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		utils.PrettyPrintJSON(security.Result{Status: "OK"})
	}
	os.Exit(0)
}

// SecurityCreateRole : Create a role in an existing Keycloak realm
func SecurityCreateRole(c *cli.Context) {
	err := security.SecRoleCreate(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		utils.PrettyPrintJSON(security.Result{Status: "OK"})
	}
	os.Exit(0)
}

// SecurityClientCreate : Create a new client in Keycloak
func SecurityClientCreate(c *cli.Context) {
	err := security.SecClientCreate(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		utils.PrettyPrintJSON(security.Result{Status: "OK"})
	}
	os.Exit(0)
}

// SecurityClientGet : Retrieve a client configuration from Keycloak
func SecurityClientGet(c *cli.Context) {
	registeredClient, err := security.SecClientGet(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if registeredClient != nil {
		utils.PrettyPrintJSON(registeredClient)
		os.Exit(0)
	}
	utils.PrettyPrintJSON(security.Result{Status: "Not found"})
	os.Exit(1)
}

// SecurityClientGetSecret : Retrieve a client secret from Keycloak
func SecurityClientGetSecret(c *cli.Context) {
	registeredClientSecret, err := security.SecClientGetSecret(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if registeredClientSecret != nil {
		utils.PrettyPrintJSON(registeredClientSecret)
		os.Exit(0)
	}
	utils.PrettyPrintJSON(security.Result{Status: "Not found"})
	os.Exit(1)
}

// SecurityUserCreate : Create a user in a Keycloak realm
func SecurityUserCreate(c *cli.Context) {
	err := security.SecUserCreate(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		utils.PrettyPrintJSON(security.Result{Status: "OK"})
	}
	os.Exit(0)
}

// SecurityUserGet : Retrieve the user detail from Keycloak
func SecurityUserGet(c *cli.Context) {
	registeredUser, err := security.SecUserGet(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if registeredUser != nil {
		utils.PrettyPrintJSON(registeredUser)
		os.Exit(0)
	}
	utils.PrettyPrintJSON(security.Result{Status: "Not found"})
	os.Exit(1)
}

// SecurityUserSetPassword : Set a users password in Keycloak
func SecurityUserSetPassword(c *cli.Context) {
	err := security.SecUserSetPW(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrintJSON(security.Result{Status: "OK"})
	os.Exit(0)
}

// SecurityUserAddRole : Add an existing role to the specified user
func SecurityUserAddRole(c *cli.Context) {
	err := security.SecUserAddRole(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		utils.PrettyPrintJSON(security.Result{Status: "OK"})
	}
	os.Exit(0)
}

// SecurityKeyUpdate : Creates or updates a key in the platforms keyring
func SecurityKeyUpdate(c *cli.Context) {
	connectionID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	username := strings.TrimSpace(strings.ToLower(c.String("username")))
	password := strings.TrimSpace(c.String("password"))
	err := security.SecKeyUpdate(connectionID, username, password)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	response, _ := json.Marshal(security.Result{Status: "OK"})
	fmt.Println(string(response))
	os.Exit(0)
}

// SecurityKeyValidate : Checks the key is available in the platform keyring
func SecurityKeyValidate(c *cli.Context) {
	connectionID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	username := strings.TrimSpace(strings.ToLower(c.String("username")))
	_, err := security.SecKeyGetSecret(connectionID, username)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	response, _ := json.Marshal(security.Result{Status: "OK"})
	fmt.Println(string(response))
	os.Exit(0)
}
