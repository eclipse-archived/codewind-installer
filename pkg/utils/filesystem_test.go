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

package utils

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO : Move these 2 helpers somewhere central, for sharing between packages

// CreateTempTestFile : creates a temp test file and returns that file and a clean-up function
func CreateTempTestFile(t *testing.T, initialData string) (*os.File, func()) {
	t.Helper()
	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatalf("could not create temp file %v", err)
	}

	tmpfile.Write([]byte(initialData))

	removeFile := func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}
	return tmpfile, removeFile
}

// CreateTempTestDir : creates a temp test dir and returns that dir and a clean-up function
func CreateTempTestDir(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("could not create temp file %v", err)
	}

	removeDir := func() {
		os.RemoveAll(tmpDir)
	}
	return tmpDir, removeDir
}

func TestCreateTempFile(t *testing.T) {
	testPath := "create_temp_file_test_delete_me"
	t.Run("success case - creates temporary file", func(*testing.T) {
		err := CreateTempFile(testPath)
		assert.Nil(t, err)

		_, err = os.Stat(testPath)
		assert.Nil(t, err)
		assert.Equal(t, PathExists(testPath), true)

		os.Remove(testPath)
	})
	t.Run("doesn't change an existing file", func(t *testing.T) {
		file, removeFile := CreateTempTestFile(t, "Hello World")
		defer removeFile()
		err := CreateTempFile(file.Name())
		assert.Nil(t, err)

		writeFileContent := []byte("Hello World")
		fileContent, _ := ioutil.ReadFile(file.Name())
		assert.Equal(t, writeFileContent, fileContent)
	})
}

func TestPathExists(t *testing.T) {
	t.Run("returns true when path exists", func(t *testing.T) {
		file, removeFile := CreateTempTestFile(t, "")
		defer removeFile()
		got := PathExists(file.Name())
		assert.True(t, got)
	})
	t.Run("returns false when path does not exist", func(t *testing.T) {
		testFile := "create_test_file_delete_me_2"
		got := PathExists(testFile)
		assert.False(t, got)
		os.Remove(testFile)
	})
}

func TestDirIsEmpty(t *testing.T) {
	t.Run("returns true when dir is empty", func(t *testing.T) {
		testDir, removeDir := CreateTempTestDir(t)
		defer removeDir()
		got, err := DirIsEmpty(testDir)

		assert.True(t, got)
		assert.Nil(t, err)
	})
	t.Run("returns false when dir is non-empty", func(t *testing.T) {
		testDir, removeDir := CreateTempTestDir(t)
		defer removeDir()
		ioutil.WriteFile(path.Join(testDir, "test"), []byte{}, 0777)
		got, err := DirIsEmpty(testDir)

		assert.False(t, got)
		assert.Nil(t, err)
	})
	t.Run("returns error when dir doesn't exist", func(t *testing.T) {
		got, err := DirIsEmpty("not-created")

		assert.False(t, got)
		assert.NotNil(t, err)
		os.Remove(testDir)
	})
}

func TestReplaceInFiles(t *testing.T) {
	t.Run("replaces placeholder in file", func(t *testing.T) {
		testFile, removeFile := CreateTempTestFile(t, "[PROJ_NAME_PLACEHOLDER] test")
		defer removeFile()
		err := ReplaceInFiles(testFile.Name(), "[PROJ_NAME_PLACEHOLDER]", "project")
		assert.Nil(t, err)

		fileContent, _ := ioutil.ReadFile(testFile.Name())
		wantFileContent := []byte("project test")
		assert.Equal(t, wantFileContent, fileContent)
	})
}
