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

// File schema versions of the Connections configuration file

package connections

// ConnectionConfigV0 : ConnectionsConfig Schema Version 0
type ConnectionConfigV0 struct {
	Connections []ConnectionV0 `json:"connections"`
}

// ConnectionConfigV1 : Connections Schema Version 1
type ConnectionConfigV1 struct {
	SchemaVersion int            `json:"schemaversion"`
	Connections   []ConnectionV1 `json:"connections"`
}

// ConnectionV0 : Connections Schema Version 0
type ConnectionV0 struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	AuthURL  string `json:"auth"`
	Realm    string `json:"realm"`
	ClientID string `json:"client_id"`
}

// ConnectionV1 : Connections Schema Version 1
type ConnectionV1 struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	AuthURL  string `json:"auth"`
	Realm    string `json:"realm"`
	ClientID string `json:"client_id"`
}
