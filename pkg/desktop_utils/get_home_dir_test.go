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

package desktoputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test to make sure GetHomeDir() does not return a nil value
func Test_GetHomeDir(t *testing.T) {
	result := GetHomeDir()
	assert.NotNil(t, result, "should return home dir of system or blank string")
}
