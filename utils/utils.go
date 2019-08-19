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

package utils

// RemoveArrayDuplicate elements
func RemoveArrayDuplicate(tagArr []string) []string {
	// Use map to record duplicates if found
	encountered := map[string]bool{}
	result := []string{}

	for element := range tagArr {
		if encountered[tagArr[element]] == true {
			// Don't add duplicate
		} else {
			// Record said element as an encountered
			encountered[tagArr[element]] = true
			// Append to the new result slice
			result = append(result, tagArr[element])
		}
	}

	return result
}
