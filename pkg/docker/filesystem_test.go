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

package docker

import (
	"os"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/globals"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestWriteToComposeFile(t *testing.T) {
	originalUseInsecureKeyring := globals.UseInsecureKeyring
	globals.SetUseInsecureKeyring(true)
	t.Run("docker compose should be written to the filepath", func(t *testing.T) {
		testFile := "TestFile.yaml"
		os.Create(testFile)
		_, err := WriteToComposeFile("TestFile.yaml", false)

		pathExists := utils.PathExists(testFile)
		assert.True(t, pathExists)
		assert.Nil(t, err)
		os.Remove(testFile)
	})

	t.Run("docker compose should be written to the filepath", func(t *testing.T) {
		testFile := "TestFile.yaml"
		os.Create(testFile)
		composeWritten, err := WriteToComposeFile("TestFile.yaml", false)

		pathExists := utils.PathExists(testFile)
		assert.True(t, pathExists)
		assert.True(t, composeWritten)
		assert.Nil(t, err)
		os.Remove(testFile)
	})
	t.Run("empty path returns nil", func(t *testing.T) {
		composeWritten, err := WriteToComposeFile("", false)
		assert.False(t, composeWritten)
		assert.Nil(t, err)
	})
	globals.SetUseInsecureKeyring(originalUseInsecureKeyring)
}
