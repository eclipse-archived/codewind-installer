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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/sechttp"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

type (
	// CompleteRequest is the request body format for calling the upload complete API
	CompleteRequest struct {
		FileList      []string `json:"fileList"`
		DirectoryList []string `json:"directoryList"`
		ModifiedList  []string `json:"modifiedList"`
		TimeStamp     int64    `json:"timeStamp"`
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

	// walkerInfo is the input struct to the walker function
	walkerInfo struct {
		Path         string   // the path of the current file
		os.FileInfo           // the FileInfo of the current file
		IgnoredPaths []string // paths to ignore
		LastSync     int64    // last sync time
	}

	// SyncInfo contains the information from a project sync
	SyncInfo struct {
		fileList         []string
		directoryList    []string
		modifiedList     []string
		UploadedFileList []UploadedFile
	}

	// refPath is a referenced file path to sync
	refPath struct {
		From string `json:"from"`
		To   string `json:"to"`
	}

	// refPaths is an array of refPath objects
	refPaths struct {
		RefPaths []refPath
	}
)

// SyncProject syncs a project with its remote connection
func SyncProject(c *cli.Context) (*SyncResponse, *ProjectError) {
	var currentSyncTime = time.Now().UnixNano() / 1000000
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
	syncInfo, syncErr := syncFiles(&http.Client{}, projectPath, projectID, conURL, synctime, conInfo)

	// Complete the upload
	completeRequest := CompleteRequest{
		FileList:      syncInfo.fileList,
		DirectoryList: syncInfo.directoryList,
		ModifiedList:  syncInfo.modifiedList,
		TimeStamp:     currentSyncTime,
	}
	completeStatus, completeStatusCode := completeUpload(&http.Client{}, projectID, completeRequest, conInfo, conURL)
	response := SyncResponse{
		UploadedFiles: syncInfo.UploadedFileList,
		Status:        completeStatus,
		StatusCode:    completeStatusCode,
	}

	return &response, syncErr
}

func syncFiles(client utils.HTTPClient, projectPath string, projectID string, conURL string, synctime int64, connection *connections.Connection) (*SyncInfo, *ProjectError) {
	var fileList []string
	var directoryList []string
	var modifiedList []string
	var uploadedFiles []UploadedFile

	projectUploadURL := conURL + "/api/v1/projects/" + projectID + "/upload"

	refPathsChanged := false

	// define a walker function
	walker := func(path string, info walkerInfo, err error) error {
		if err != nil {
			return err
			// TODO - How to handle *some* files being unreadable
		}

		// If it is the top level directory ignore it
		if path == projectPath {
			return nil
		}

		// use ToSlash to try and get both Windows and *NIX paths to be *NIX for pfe
		relativePath := filepath.ToSlash(path[(len(projectPath) + 1):])

		if !info.IsDir() {
			shouldIgnore := ignoreFileOrDirectory(relativePath, false, info.IgnoredPaths)
			if shouldIgnore {
				return nil
			}
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
			if modifiedmillis > info.LastSync {
				fileContent, err := ioutil.ReadFile(info.Path)
				// Skip this file if there is an error reading it.
				if err != nil {
					return nil
				}
				// Create list of all modfied files
				modifiedList = append(modifiedList, relativePath)

				var buffer bytes.Buffer
				zWriter := zlib.NewWriter(&buffer)
				zWriter.Write([]byte(fileContent))

				zWriter.Close()
				encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())
				fileUploadBody.Message = encoded

				buf := new(bytes.Buffer)
				json.NewEncoder(buf).Encode(fileUploadBody)

				// TODO - How do we handle partial success?
				request, err := http.NewRequest("PUT", projectUploadURL, bytes.NewReader(buf.Bytes()))
				request.Header.Set("Content-Type", "application/json")
				resp, httpSecError := sechttp.DispatchHTTPRequest(client, request, connection)
				uploadedFiles = append(uploadedFiles, UploadedFile{
					FilePath:   relativePath,
					Status:     resp.Status,
					StatusCode: resp.StatusCode,
				})
				if httpSecError != nil {
					return nil
				}
				defer resp.Body.Close()

				// if this file changed, it should force referenced files to re-sync
				if relativePath == ".cw-refpaths.json" {
					refPathsChanged = true
				}
			}
		} else {
			shouldIgnore := ignoreFileOrDirectory(relativePath, true, info.IgnoredPaths)
			if shouldIgnore {
				return filepath.SkipDir
			}
			directoryList = append(directoryList, relativePath)
		}
		return nil
	}

	// read the ignored and referenced paths into lists
	cwSettingsIgnoredPathsList := retrieveIgnoredPathsList(projectPath)
	cwRefPathsList := retrieveRefPathsList(projectPath)

	// initialize a combined list, prime it with ignored paths from .cw-settings
	// then append with referenced "To" paths
	cwCombinedIgnoredPathsList := append([]string{}, cwSettingsIgnoredPathsList...)
	for _, refPath := range cwRefPathsList {
		cwCombinedIgnoredPathsList = append(cwCombinedIgnoredPathsList, refPath.To)
	}

	// first sync files that are physically in the project
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		// use combined ignored paths here, files in the project that
		// are also the target of a reference should not be synced
		wInfo := walkerInfo{
			path,
			info,
			cwCombinedIgnoredPathsList,
			synctime,
		}
		return walker(path, wInfo, err)
	})
	if err != nil {
		text := fmt.Sprintf("error walking the path %q: %v\n", projectPath, err)
		return nil, &ProjectError{errOpSync, errors.New(text), text}
	}

	errText := ""

	// then sync referenced file paths
	for _, refPath := range cwRefPathsList {

		// get From path and resolve to absolute if needed
		from := refPath.From
		if !filepath.IsAbs(from) {
			from = filepath.Join(projectPath, from)
		}

		// get info on the referenced file; skip invalid paths
		info, err := os.Stat(from)
		if err != nil || info.IsDir() {
			text := fmt.Sprintf("invalid file reference %q: %v\n", from, err)
			errText += text
			continue
		}

		lastSync := synctime
		// force re-sync if .cw-refpaths.json itself was changed
		if refPathsChanged {
			lastSync = 0
		}

		// now pass it to the walker function
		wInfo := walkerInfo{
			from,
			info,
			cwSettingsIgnoredPathsList,
			lastSync,
		}
		// "To" path is relative to the project
		walker(filepath.Join(projectPath, refPath.To), wInfo, nil)
	}

	if errText != "" {
		return &SyncInfo{fileList, directoryList, modifiedList, uploadedFiles}, &ProjectError{errOpSyncRef, errors.New(errText), errText}
	}

	return &SyncInfo{fileList, directoryList, modifiedList, uploadedFiles}, nil
}

func completeUpload(client utils.HTTPClient, projectID string, completeRequest CompleteRequest, conInfo *connections.Connection, conURL string) (string, int) {
	uploadEndURL := conURL + "/api/v1/projects/" + projectID + "/upload/end"
	jsonPayload, _ := json.Marshal(&completeRequest)
	req, err := http.NewRequest("POST", uploadEndURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("error creating request %v\n", err)
		return err.Error(), 0
	}

	req.Header.Set("Content-Type", "application/json")
	resp, httpSecError := sechttp.DispatchHTTPRequest(client, req, conInfo)
	if httpSecError != nil {
		fmt.Printf("error making request  %v\n", httpSecError)
		return httpSecError.Desc, 0
	}
	defer resp.Body.Close()

	return resp.Status, resp.StatusCode
}

// Retrieve the ignoredPaths list from a .cw-settings file
func retrieveIgnoredPathsList(projectPath string) []string {
	cwSettingsPath := filepath.Join(projectPath, ".cw-settings")
	var cwSettingsIgnoredPathsList []string
	if _, err := os.Stat(cwSettingsPath); !os.IsNotExist(err) {
		plan, _ := ioutil.ReadFile(cwSettingsPath)
		var cwSettingsJSON CWSettings
		err = json.Unmarshal(plan, &cwSettingsJSON)
		if err == nil {
			cwSettingsIgnoredPathsList = cwSettingsJSON.IgnoredPaths
		}
	}
	return cwSettingsIgnoredPathsList
}

// Retrieve the refPaths list from a .cw-refpaths.json file
func retrieveRefPathsList(projectPath string) []refPath {
	cwRefPathsPath := filepath.Join(projectPath, ".cw-refpaths.json")
	var cwRefPathsList []refPath
	if _, err := os.Stat(cwRefPathsPath); !os.IsNotExist(err) {
		plan, _ := ioutil.ReadFile(cwRefPathsPath)
		var cwRefPathsJSON refPaths
		err = json.Unmarshal(plan, &cwRefPathsJSON)
		if err == nil {
			cwRefPathsList = cwRefPathsJSON.RefPaths
		}
	}
	return cwRefPathsList
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
		"*.iml",
		".project",
		".classpath",
		".options",
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
		".idea",
		".vscode",
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
		fileName = filepath.Clean(fileName)
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
