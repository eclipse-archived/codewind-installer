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

import (
	"os"
	"testing"
)

func TestCreateTempFile(t *testing.T) {
	file := CreateTempFile("TestFile.yaml")
	if file != true {
		t.Errorf("Testfile.yaml failed to create")
	}

	os.Remove("./TestFile.yaml")
}

func TestWriteToComposeFile(t *testing.T) {
	os.Create("TestFile.yaml")
	got := WriteToComposeFile("TestFile.yaml")
	if got != true {
		t.Errorf("Failed to write data to TestFile.yaml")
	}

	os.Remove("TestFile.yaml")
}

func TestDeleteTempFile(t *testing.T) {
	os.Create("TestFile.yaml")
	removeFile := DeleteTempFile("TestFile.yaml")
	if removeFile != true {
		t.Errorf("Failed to delete TestFile.yaml")
	}

}
