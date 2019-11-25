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
	"errors"
	"flag"
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// SetupKeycloak : sets up keycloak with a realm, client and user
func SetupKeycloak(codewindInstance Codewind, deployOptions *DeployOptions) error {

	accessRoleName := "codewind-" + codewindInstance.WorkspaceID

	authURL := KeycloakPrefix + codewindInstance.Ingress
	if deployOptions.KeycloakTLSSecure {
		authURL = "https://" + authURL
	} else {
		authURL = "http://" + authURL
	}
	gateKeeperProtocol := "http://"
	if deployOptions.GateKeeperTLSSecure {
		gateKeeperProtocol = "https://"
	}

	logr.Infoln("Waiting for Keycloak to start")
	startErr := utils.WaitForService(authURL, 200, 500)
	if startErr != nil {
		return errors.New("Keycloak did not start in a reasonable about of time")
	}

	logr.Infoln("Configuring Keycloak...")
	logr.Infoln(authURL)

	// Authenticate to get an admin access token
	flagSet := flag.NewFlagSet("authentication", 0)
	flagSet.String("host", authURL, "doc")
	flagSet.String("realm", "master", "doc")
	flagSet.String("username", deployOptions.KeycloakUser, "doc")
	flagSet.String("password", deployOptions.KeycloakPassword, "doc")
	flagSet.String("client", "admin-cli", "doc")
	c := cli.NewContext(nil, flagSet, nil)
	tokens, secErr := security.SecAuthenticate(http.DefaultClient, c, "", "")
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	// Create a new realm
	logr.Infoln("Creating Keycloak realm")
	realmFlagset := flag.NewFlagSet("setupRealm", 0)
	realmFlagset.String("host", authURL, "doc")
	realmFlagset.String("newrealm", deployOptions.KeycloakRealm, "doc")
	realmFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c = cli.NewContext(nil, realmFlagset, nil)
	secErr = security.SecRealmCreate(c)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	// Create a new client
	logr.Infoln("Creating Keycloak client")
	gatekeeperPublicURL := gateKeeperProtocol + GatekeeperPrefix + codewindInstance.Ingress + "/*"
	clientFlagset := flag.NewFlagSet("setupClient", 0)
	clientFlagset.String("host", authURL, "doc")
	clientFlagset.String("redirect", gatekeeperPublicURL, "doc")
	clientFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	clientFlagset.String("newclient", deployOptions.KeycloakClient, "doc")
	clientFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c = cli.NewContext(nil, clientFlagset, nil)
	secErr = security.SecClientCreate(c)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	// Create a new access role for this deployment
	logr.Infof("Creating access role '%v' in realm '%v'", accessRoleName, deployOptions.KeycloakRealm)
	clientFlagset = flag.NewFlagSet("setupClient", 0)
	clientFlagset.String("host", authURL, "doc")
	clientFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	clientFlagset.String("role", accessRoleName, "doc")
	clientFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c = cli.NewContext(nil, clientFlagset, nil)
	secErr = security.SecRoleCreate(c)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	// Create an initial user
	logr.Infoln("Creating Keycloak initial user")
	userCreateFlagset := flag.NewFlagSet("createUser", 0)
	userCreateFlagset.String("host", authURL, "doc")
	userCreateFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	userCreateFlagset.String("name", deployOptions.KeycloakDevUser, "doc")
	userCreateFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c = cli.NewContext(nil, userCreateFlagset, nil)
	secErr = security.SecUserCreate(c)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	// Create an initial user password
	logr.Infoln("Updating Keycloak user password")
	userPassFlagset := flag.NewFlagSet("createUser", 0)
	userPassFlagset.String("host", authURL, "doc")
	userPassFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	userPassFlagset.String("name", deployOptions.KeycloakDevUser, "doc")
	userPassFlagset.String("newpw", deployOptions.KeycloakDevPassword, "doc")
	userPassFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c = cli.NewContext(nil, userPassFlagset, nil)
	secErr = security.SecUserSetPW(c)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	// Grant the user access to this Deployment
	logr.Printf("Grant '%v' access to this deployment ", deployOptions.KeycloakDevUser)
	clientFlagset = flag.NewFlagSet("setupClient", 0)
	clientFlagset.String("host", authURL, "doc")
	clientFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	clientFlagset.String("role", accessRoleName, "doc")
	clientFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c = cli.NewContext(nil, clientFlagset, nil)
	secErr = security.SecUserAddRole(c)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	// Load client secret
	logr.Infoln("Fetching client secret")
	clientSecFlagset := flag.NewFlagSet("getClientSecret", 0)
	clientSecFlagset.String("host", authURL, "doc")
	clientSecFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	clientSecFlagset.String("clientid", deployOptions.KeycloakClient, "doc")
	clientSecFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c = cli.NewContext(nil, clientSecFlagset, nil)
	registeredSecret, secErr := security.SecClientGetSecret(c)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}
	deployOptions.ClientSecret = registeredSecret.Secret
	return nil
}
