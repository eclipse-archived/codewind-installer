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

package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_ProjectConnection :  Tests
func Test_ProjectConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	t.Run("Asserts no connectionID is found for non-existant project", func(t *testing.T) {
		connectionID, projError := GetConnectionID("1234-abcd")
		if projError != nil {
			t.Fail()
		}
		assert.Equal(t, "", connectionID)
	})
}
