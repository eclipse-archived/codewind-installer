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
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/urfave/cli"
)

type (
	// CompleteRequest is the request body format for calling the upload complete API
	CompleteRequest struct {
		FileList     []string `json:"fileList"`
		ModifiedList []string `json:"modifiedList"`
		TimeStamp    int64    `json:"timeStamp"`
	}

	// FileUploadMsg is the message sent on uploading a file
	FileUploadMsg struct {
		IsDirectory  bool   `json:"isDirectory"`
		Mode         uint   `json:"mode"`
		RelativePath string `json:"path"`
		Message      string `json:"msg"`
	}

	// UploadedFile is the file to sync
	UploadedFile struct {
		FilePath   string `json:"filePath"`
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
	}

	// SyncResponse is the status of the file syncing
	SyncResponse struct {
		Status        string         `json:"status"`
		StatusCode    int            `json:"statusCode"`
		UploadedFiles []UploadedFile `json:"uploadedFiles"`
	}
)

// SyncProject syncs a project with its remote connection
func SyncProject(c *cli.Context) (*SyncResponse, *ProjectError) {
	projectPath := strings.TrimSpace(c.String("path"))
	projectID := strings.TrimSpace(c.String("id"))
	synctime := int64(c.Int("time"))
	_, err := os.Stat(projectPath)
	if err != nil {
		return nil, &ProjectError{errBadPath, err, err.Error()}
	}

	if !ConnectionFileExists(projectID) {
		fmt.Println("Project connection file does not exist, creating default local connection")
		CreateConnectionFile(projectID)
	}

	conID, projErr := GetConnectionID(projectID)

	if projErr != nil {
		return nil, projErr
	}

	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return nil, &ProjectError{errOpConNotFound, conInfoErr, conInfoErr.Desc}
	}

	conURL, conURLErr := config.PFEOriginFromConnection(conInfo)
	if conURLErr != nil {
		return nil, &ProjectError{errOpConNotFound, conURLErr.Err, conURLErr.Desc}
	}

	// Sync all the necessary project files
	fileList, modifiedList, uploadedFilesList := syncFiles(projectPath, projectID, conURL, synctime, conInfo)
	// Complete the upload
	completeStatus, completeStatusCode := completeUpload(projectID, fileList, modifiedList, conURL, synctime)
	response := SyncResponse{
		UploadedFiles: uploadedFilesList,
		Status:        completeStatus,
		StatusCode:    completeStatusCode,
	}

	return &response, nil
}

func syncFiles(projectPath string, projectID string, conURL string, synctime int64, connection *connections.Connection) ([]string, []string, []UploadedFile) {
	var fileList []string
	var modifiedList []string
	var uploadedFiles []UploadedFile

	projectUploadURL := conURL + "/api/v1/projects/" + projectID + "/upload"
	client := &http.Client{}

	cwSettingsIgnoredPathsList := retrieveIgnoredPathsList(projectPath)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			panic(err)
			// TODO - How to handle *some* files being unreadable
		}

		if !info.IsDir() {
			shouldIgnore := ignoreFileOrDirectory(info.Name(), false, cwSettingsIgnoredPathsList)
			if shouldIgnore {
				return nil
			}
			// use ToSlash to try and get both Windows and *NIX paths to be *NIX for pfe
			relativePath := filepath.ToSlash(path[(len(projectPath) + 1):])
			// Create list of all files for a project
			fileList = append(fileList, relativePath)

			// get time file was modified in milliseconds since epoch
			modifiedmillis := info.ModTime().UnixNano() / 1000000

			fileUploadBody := FileUploadMsg{
				IsDirectory:  info.IsDir(),
				Mode:         uint(info.Mode().Perm()),
				RelativePath: relativePath,
				Message:      "",
			}

			// Has this file been modified since last sync
			if modifiedmillis > synctime {
				fileContent, err := ioutil.ReadFile(path)
				jsonContent, err := json.Marshal(string(fileContent))
				// Skip this file if there is an error reading it.
				if err != nil {
					return nil
				}
				// Create list of all modfied files
				modifiedList = append(modifiedList, relativePath)

				var buffer bytes.Buffer
				zWriter := zlib.NewWriter(&buffer)
				zWriter.Write([]byte(jsonContent))

				zWriter.Close()
				encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())
				fileUploadBody.Message = encoded

				buf := new(bytes.Buffer)
				json.NewEncoder(buf).Encode(fileUploadBody)

				// TODO - How do we handle partial success?
				request, err := http.NewRequest("PUT", projectUploadURL, bytes.NewReader(buf.Bytes()))
				request.Header.Set("Content-Type", "application/json")
				resp, httpSecError := sechttp.DispatchHTTPRequest(client, request, connection.Username, connection.ID)
				uploadedFiles = append(uploadedFiles, UploadedFile{
					FilePath:   relativePath,
					Status:     resp.Status,
					StatusCode: resp.StatusCode,
				})
				if httpSecError != nil {
					return nil
				}
				defer resp.Body.Close()
			}
		} else {
			shouldIgnore := ignoreFileOrDirectory(info.Name(), true, cwSettingsIgnoredPathsList)
			if shouldIgnore {
				return filepath.SkipDir
			}

		}

		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", projectPath, err)
		return nil, nil, nil
	}
	return fileList, modifiedList, uploadedFiles
}

func completeUpload(projectID string, files []string, modfiles []string, conURL string, timestamp int64) (string, int) {
	uploadEndURL := conURL + "/api/v1/projects/" + projectID + "/upload/end"

	payload := &CompleteRequest{FileList: files, ModifiedList: modfiles, TimeStamp: timestamp}
	jsonPayload, _ := json.Marshal(payload)

	// Make the request to end the sync process.
	resp, err := http.Post(uploadEndURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		panic(err)
		// TODO - Need to handle this gracefully.
	}
	return resp.Status, resp.StatusCode
}

// Retrieve the ignoredPaths list from a .cw-settings file
func retrieveIgnoredPathsList(projectPath string) []string {
	cwSettingsPath := path.Join(projectPath, ".cw-settings")
	var cwSettingsIgnoredPathsList []string
	if _, err := os.Stat(cwSettingsPath); !os.IsNotExist(err) {
		plan, _ := ioutil.ReadFile(cwSettingsPath)
		var cwSettingsJSON CWSettings
		// Don't need to handle an invalid JSON file as we should just return []
		json.Unmarshal(plan, &cwSettingsJSON)
		cwSettingsIgnoredPathsList = cwSettingsJSON.IgnoredPaths
	}
	return cwSettingsIgnoredPathsList
}

func ignoreFileOrDirectory(name string, isDir bool, cwSettingsIgnoredPathsList []string) bool {
	// List of files that will not be sent to PFE
	ignoredFiles := []string{
		".DS_Store",
		"*.swp",
		"*.swx",
		"Jenkinsfile",
		".cfignore",
		"localm2cache.zip",
		"libertyrepocache.zip",
		"run-dev",
		"run-debug",
		"manifest.yml",
		"idt.js",
		".bluemix",
		".build-ubuntu",
		".yo-rc.json",
	}

	// List of directories that will not be sent to PFE
	ignoredDirectories := []string{
		".project",
		"node_modules*",
		".git*",
		"load-test*",
		".settings",
		"Dockerfile-tools",
		"target",
		"mc-target",
		".m2",
		"debian",
		".bluemix",
		"terraform",
		".build-ubuntu",
	}

	ignoredList := ignoredFiles
	if isDir {
		ignoredList = ignoredDirectories
	}

	if len(cwSettingsIgnoredPathsList) > 0 {
		ignoredList = append(ignoredList, cwSettingsIgnoredPathsList...)
	}

	isFileInIgnoredList := false
	for _, fileName := range ignoredList {
		matched, err := filepath.Match(fileName, name)
		if err != nil {
			return false
		}
		if matched {
			isFileInIgnoredList = true
			break
		}
	}
	return isFileInIgnoredList
}

// PrettyPrintJSON : Format JSON output for display
func PrettyPrintJSON(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}
