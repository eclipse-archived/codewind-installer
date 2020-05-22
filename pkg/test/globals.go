/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

package test

// ConID is the default connection ID
const ConID = "local"

// PublicGHRepoURL is a URL to a public GitHub repo (not requiring auth to access)
const PublicGHRepoURL = "https://github.com/microclimate-dev2ops/nodeExpressTemplate"

// PublicGHDevfileURL is a URL to a devfiles/index.json in a public GitHub repo (not requiring auth to access)
const PublicGHDevfileURL = "https://raw.githubusercontent.com/kabanero-io/codewind-templates/aad4bafc14e1a295fb8e462c20fe8627248609a3/devfiles/index.json"

// GHERepoURL is a URL to a GitHub Enterprise repo (requiring auth to access)
const GHERepoURL = "https://github.ibm.com/DevCamp2018/git-basics"

// GHEDevfileURL is a URL to a devfiles/index.json in a GitHub Enterprise repo
const GHEDevfileURL = "https://raw.github.ibm.com/Richard-Waller/sampleGHETemplateRepo/2a0cb5e1accfc077e375f6582f7ee27015c4005b/devfiles/index.json"

// GHEUsername is a username that passes the auth required to access a GHERepoURL
const GHEUsername = "INSERT YOUR OWN: e.g. foo.bar@foobar.com"

// GHEPassword is a password that passes the auth required to access a GHERepoURL
const GHEPassword = "INSERT YOUR OWN: e.g. 1234kljfdsjfaleru29348spodkfj445"

// GHEPersonalAccessToken is a personal access token that passes the auth required to access a GHERepoURL
const GHEPersonalAccessToken = "INSERT YOUR OWN"

// UsingOwnGHECredentials should be set to true if you want to run tests using the credentials above
const UsingOwnGHECredentials = false
