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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/eclipse/codewind-installer/pkg/errors"
	"github.com/google/go-github/github"
	"gopkg.in/yaml.v3"
)

// CreateTempFile in the same directory as the binary for docker compose
func CreateTempFile(filePath string) bool {
	var _, err = os.Stat(filePath)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(filePath)
		errors.CheckErr(err, 201, "")
		defer file.Close()

		dir, _ := os.Getwd()
		fmt.Println("==> created file", path.Join(dir, filePath))
		return true
	}
	return false
}

// WriteToComposeFile the contents of the docker compose yaml
func WriteToComposeFile(dockerComposeFile string, debug bool) bool {
	if dockerComposeFile == "" {
		return false
	}

	dataStruct := Compose{}

	unmarshDataErr := yaml.Unmarshal([]byte(data), &dataStruct)
	errors.CheckErr(unmarshDataErr, 202, "")

	if debug == true && len(dataStruct.SERVICES.PFE.Ports) > 0 {
		debugPort := DetermineDebugPortForPFE()
		// Add the debug port to the docker compose data
		dataStruct.SERVICES.PFE.Ports = append(dataStruct.SERVICES.PFE.Ports, "127.0.0.1:"+debugPort+":9777")
	}

	marshalledData, err := yaml.Marshal(&dataStruct)
	errors.CheckErr(err, 203, "")

	if debug == true {
		fmt.Printf("==> %s structure is: \n%s\n\n", dockerComposeFile, string(marshalledData))
	} else {
		fmt.Println("==> environment structure written to " + dockerComposeFile)
	}

	err = ioutil.WriteFile(dockerComposeFile, marshalledData, 0644)
	errors.CheckErr(err, 204, "")
	return true
}

// DeleteTempFile once the the Codewind environment has been created
func DeleteTempFile(filePath string) (bool, error) {
	var _, file = os.Stat(filePath)

	if os.IsNotExist(file) {
		errors.CheckErr(file, 206, "No files to delete")
		return false, file
	}

	os.Remove(filePath)
	// fmt.Printf("==> Deleted file: %s\n", filePath)
	return true, nil
}

// PingHealth - pings environment api every 15 seconds to check if containers started
func PingHealth(healthEndpoint string) bool {
	var started = false
	fmt.Println("Waiting for Codewind to start")
	hostname, port := GetPFEHostAndPort()
	for i := 0; i < 120; i++ {
		resp, err := http.Get("http://" + hostname + ":" + port + healthEndpoint)
		if err != nil {
			fmt.Printf(".")
		} else {
			if resp.StatusCode == 200 {
				fmt.Println("\nHTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
				fmt.Println("Codewind successfully started on http://" + hostname + ":" + port)
				started = true
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	if started != true {
		log.Fatal("Codewind containers are taking a while to start. Please check the container logs and/or restart Codewind")
	}
	return started
}

// GetZipURL from github api /repos/:owner/:repo/:archive_format/:ref
func GetZipURL(owner, repo, branch string) (string, error) {
	client := github.NewClient(nil)

	opt := &github.RepositoryContentGetOptions{Ref: branch}

	URL, _, err := client.Repositories.GetArchiveLink(context.Background(), owner, repo, "zipball", opt, true)
	if err != nil {
		return "", err
	}
	url := URL.String()
	return url, nil
}

// DownloadFile from URL to file destination
func DownloadFile(URL, destination string) error {
	// Get the data
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	file, err := os.Create(destination)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	// Write body to file
	_, err = io.Copy(file, resp.Body)
	log.Printf("Downloaded file from '%s' to '%s'\n", URL, destination)

	return err
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

		fileNameArr := strings.Split(file.Name, "/")
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
	log.Printf("Extracted file from '%s' to '%s'\n", filePath, destination)
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
			fileToOverwrite, err := overwriteFile(target)
			defer fileToOverwrite.Close()
			if err != nil {
				log.Fatal(err)
			}
			if _, err := io.Copy(fileToOverwrite, tarReader); err != nil {
				log.Fatal(err)
			}
		default:
			log.Printf("Can't extract to %s: unknown typeflag %c\n", target, header.Typeflag)
		}
	}
	return nil
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
