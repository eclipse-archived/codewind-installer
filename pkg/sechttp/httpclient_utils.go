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

package sechttp

const (
	errOpNoConnection = "tx_connection"
	errOpAuthFailed   = "tx_auth"
	errOpFailed       = "tx_failed"
	errOpNoPassword   = "tx_nopassword"
)

const (
	errConnetionNotFound = "Cant find a valid connection"
	errMissingPassword   = "Unable to find password in keychain"
)
