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
	"path/filepath"
	"strings"

	"github.com/eclipse/codewind-installer/config"
	"github.com/eclipse/codewind-installer/utils/deployments"
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
		RelativePath string `json:"path"`
		Message      string `json:"msg"`
	}
)

// SyncProject syncs a project with its remote deployment
func SyncProject(c *cli.Context) *ProjectError {
	projectPath := strings.TrimSpace(c.String("path"))
	projectID := strings.TrimSpace(c.String("id"))
	synctime := int64(c.Int("time"))
	depID := strings.TrimSpace(c.String("d"))

	_, err := os.Stat(projectPath)
	if err != nil {
		return &ProjectError{errBadPath, err, err.Error()}
	}

	deploymentInfo, err := deployments.GetDeploymentByID(depID)

	if err != nil {
		return &ProjectError{errOpNotFound, err, err.Error()}
	}

	var depURL string
	if depID != "local" {
		depURL = deploymentInfo.URL
	} else {
		depURL = config.PFEApiRoute()
	}

	// Sync all the necessary project files
	fileList, modifiedList := syncFiles(projectPath, projectID, depURL, synctime)

	// Complete the upload
	completeUpload(projectID, fileList, modifiedList, depURL, synctime)
	return nil
}

func syncFiles(projectPath string, projectID string, depURL string, synctime int64) ([]string, []string) {
	var fileList []string
	var modifiedList []string

	projectUploadURL := depURL + "projects/" + projectID + "/upload"
	client := &http.Client{}
	//	fmt.Println("Uploading to " + projectUploadUrl)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			panic(err)
			// TODO - How to handle *some* files being unreadable
		}
		if !info.IsDir() {
			relativePath := path[(len(projectPath) + 1):]
			// Create list of all files for a project
			fileList = append(fileList, relativePath)

			// get time file was modified in milliseconds since epoch
			modifiedmillis := info.ModTime().UnixNano() / 1000000

			fileUploadBody := FileUploadMsg{
				IsDirectory:  info.IsDir(),
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
				resp, err := client.Do(request)
				fmt.Println("Upload status:" + resp.Status + " for file: " + relativePath)
				if err != nil {
					return nil
				}
			}
		}

		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", projectPath, err)
		return nil, nil
	}
	return fileList, modifiedList
}

func completeUpload(projectID string, files []string, modfiles []string, depURL string, timestamp int64) {
	uploadEndURL := depURL + "projects/" + projectID + "/upload/end"

	payload := &CompleteRequest{FileList: files, ModifiedList: modfiles, TimeStamp: timestamp}
	jsonPayload, _ := json.Marshal(payload)

	// Make the request to end the sync process.
	resp, err := http.Post(uploadEndURL, "application/json", bytes.NewBuffer(jsonPayload))
	fmt.Println("Upload end status:" + resp.Status)
	if err != nil {
		panic(err)
		// TODO - Need to handle this gracefully.
	}
}
