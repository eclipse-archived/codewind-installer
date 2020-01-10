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

package connections

const (
	errOpFileParse    = "con_parse"
	errOpFileLoad     = "con_load"
	errOpFileWrite    = "con_write"
	errOpSchemaUpdate = "con_schema_update"
	errOpConflict     = "con_conflict"
	errOpNotFound     = "con_not_found"
	errOpProtected    = "con_protected"
	errOpGetEnv       = "con_environment"
)

const (
	errTargetNotFound = "Target connection not found"
)

// Result : status message
type Result struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
}
