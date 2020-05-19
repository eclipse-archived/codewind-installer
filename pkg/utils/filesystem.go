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
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/errors"
	logr "github.com/sirupsen/logrus"
)

// CreateTempFile in the same directory as the binary for docker compose
func CreateTempFile(filePath string) error {
	var _, err = os.Stat(filePath)

	// create file if not exists
	if os.IsNotExist(err) {
		file, createErr := os.Create(filePath)
		if createErr != nil {
			return createErr
		}
		defer file.Close()
	}
	return nil
}

// UnZip unzips a file to a destination
func UnZip(filePath, destination string) error {
	zipReader, _ := zip.OpenReader(filePath)
	if zipReader == nil {
		return fmt.Errorf("file '%s' is empty", filePath)
	}

	var extractedFilePath string
	zipFiles := zipReader.Reader.File
	for _, file := range zipFiles {

		zippedFile, err := file.Open()
		errors.CheckErr(err, 402, "")
		defer zippedFile.Close()

		fileNameArr := strings.Split(file.Name, string(os.PathSeparator))
		extractedFilePath = destination

		for i := 1; i < len(fileNameArr); i++ {
			extractedFilePath = filepath.Join(extractedFilePath, fileNameArr[i])
		}

		if file.FileInfo().IsDir() {
			// For debug:
			// fmt.Println("Directory Created:", extractedFilePath)
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			// For debug:
			// fmt.Println("File extracted:", file.Name)

			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			errors.CheckErr(err, 403, "")
			defer outputFile.Close()

			_, err = io.Copy(outputFile, zippedFile)
			errors.CheckErr(err, 404, "")
		}
	}
	logr.Tracef("Extracted file from '%s' to '%s'\n", filePath, destination)
	return nil
}

// UnTar unpacks a tar.gz file to a destination
func UnTar(pathToTarFile, destination string) error {
	fileReader, err := readFile(pathToTarFile)
	if err != nil {
		return err
	}
	defer fileReader.Close()
	gzipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	return ExtractTarToFileSystem(tarReader, destination)
}

// ExtractTarToFileSystem reads a tar Reader to a filesystem
func ExtractTarToFileSystem(tarReader *tar.Reader, destination string) error {
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		target := filepath.Join(destination, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				log.Fatal(err)
			}
		case tar.TypeReg:
			extractFile(target, tarReader, header)
		default:
			log.Printf("Can't extract to %s: unknown typeflag %c\n", target, header.Typeflag)
		}
	}
	return nil
}

func extractFile(target string, tarReader *tar.Reader, header *tar.Header) {
	fileToOverwrite, err := overwriteFile(target)
	defer fileToOverwrite.Close()
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
	if _, err := io.Copy(fileToOverwrite, tarReader); err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
	os.Chmod(target, os.FileMode(header.Mode))
}

func overwriteFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0777) // gives everyone rwx permission
	if err != nil {
		file, err = os.Create(filePath)
		if err != nil {
			return file, err
		}
	}
	return file, nil
}

func readFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0444) // gives everyone read permission
	if err != nil {
		return file, err
	}
	return file, nil
}

// PathExists returns whether a path exists on the local file system.
func PathExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// DirIsEmpty returns true if the directory at the given path if empty
func DirIsEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// ReplaceInFiles the placeholder string "[PROJ_NAME_PLACEHOLDER]" with a generated name based on the project directory
func ReplaceInFiles(projectPath string, oldStr string, newStr string) error {

	oldBytes := []byte(oldStr)
	newBytes := []byte(newStr)

	pathsToRename := []string{}

	lastError := error(nil)
	filepath.Walk(projectPath, func(pathName string, info os.FileInfo, err error) error {

		if strings.Contains(path.Base(pathName), oldStr) {
			// Keep track of files we need to rename but don't rename
			// them until the filepath.Walk is complete.
			pathsToRename = append(pathsToRename, pathName)
		}

		if info.IsDir() {
			return nil
		}

		content, err := ioutil.ReadFile(pathName)
		if err != nil {
			lastError = err
			return nil
		}
		newContent := bytes.Replace(content, []byte(oldBytes), []byte(newBytes), -1)
		if err = ioutil.WriteFile(pathName, newContent, info.Mode()); err != nil {
			lastError = err
			return nil
		}
		return nil
	})

	for _, pathName := range pathsToRename {
		newPath := strings.Replace(pathName, oldStr, newStr, -1)
		os.Rename(pathName, newPath)
	}

	return lastError
}

//CopyFile - copies the contents of the source file to a target file
func CopyFile(sourceFilePath, targetFilePath string) error {
	sourceFileStat, err := os.Stat(sourceFilePath)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", sourceFilePath)
	}

	source, err := os.Open(sourceFilePath)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

//Zip - creates a zip file in the target directory and populates it with the contents of that directory
func Zip(zipFileName, targetDirectory string) error {
	newZipFile, zipCreateErr := os.Create(filepath.Join(targetDirectory, zipFileName))
	if zipCreateErr != nil {
		return fmt.Errorf("Unable to create zip file - " + zipCreateErr.Error())
	} else {
		defer newZipFile.Close()

		zipWriter := zip.NewWriter(newZipFile)
		defer zipWriter.Close()

		// Add files to zip
		err := filepath.Walk(targetDirectory, func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsRegular() && (info.Name() != zipFileName) {
				fileToZip, err := os.Open(path)
				if err != nil {
					return err
				}
				defer fileToZip.Close()

				// Get the file information
				info, err := fileToZip.Stat()
				if err != nil {
					return err
				}

				header, err := zip.FileInfoHeader(info)
				if err != nil {
					return err
				}

				// Using FileInfoHeader() above only uses the basename of the file. If we want
				// to preserve the folder structure we can overwrite this with the full path.
				header.Name = strings.Replace(path, targetDirectory+string(os.PathSeparator), "", 1)

				// Change to deflate to gain better compression
				// see http://golang.org/pkg/archive/zip/#pkg-constants
				header.Method = zip.Deflate

				writer, err := zipWriter.CreateHeader(header)
				if err != nil {
					return err
				}
				_, err = io.Copy(writer, fileToZip)
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("walk error " + err.Error())
		}
	}
	return nil
}
