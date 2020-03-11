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

package globals

// UseInsecureKeyring decides whether we should use the insecure keyring or the (secure) system keyring
var UseInsecureKeyring = false

// SetUseInsecureKeyring sets useInsecureKeyring
func SetUseInsecureKeyring(newUseInsecureKeyring bool) {
	UseInsecureKeyring = newUseInsecureKeyring
}
