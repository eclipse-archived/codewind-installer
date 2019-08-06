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

package actions

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eclipse/codewind-installer/errors"
	"github.com/eclipse/codewind-installer/utils"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
)

// CloneTemplate from github
func CloneTemplate(c *cli.Context) {
	//TODO Use go-git to do this in future
	destination := c.String("destination")

	zipURL := GetZipURL(c)
	tempName := "/tmp/test_" + time.Now().Format(time.RFC3339)
	zipFileName := tempName + ".zip"

	// download files in zip format
	if err := DownloadFile(zipFileName, zipURL); err != nil {
		log.Fatal(err)
	}

	// unzip into /tmp dir
	extracedFilePath := UnZip(zipFileName, "/tmp")

	// get top level dir from unzipped file and move files to new destination
	if strings.HasPrefix(extracedFilePath, "/tmp/") {
		extracedFilePath = strings.TrimPrefix(extracedFilePath, "/tmp/")
	}
	filePath := strings.Split(extracedFilePath, "/")
	source := "/tmp/" + filePath[0]
	MoveFiles(source, destination)

	//delete zip file
	utils.DeleteTempFile(zipFileName)

}

//GetZipURL from github api /repos/:owner/:repo/:archive_format/:ref
func GetZipURL(c *cli.Context) string {
	branch := c.String("branch")
	owner := c.String("owner")
	repo := c.String("repo")

	client := github.NewClient(nil)

	opt := &github.RepositoryContentGetOptions{Ref: branch}

	URL, _, err := client.Repositories.GetArchiveLink(context.Background(), owner, repo, "zipball", opt)
	if err != nil {
		log.Fatal(err)
	}
	url := URL.String()
	fmt.Println(URL)
	return url
}

//DownloadFile - the git zip of master into current dir
func DownloadFile(zipFileName, url string) error {

	// Get the data
	resp, err := http.Get(url)
	errors.CheckErr(err, 400, "")
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(zipFileName)
	errors.CheckErr(err, 401, "")
	defer out.Close()

	// Write body to file
	_, err = io.Copy(out, resp.Body)
	fmt.Println(zipFileName)

	return err
}

//UnZip downloaded file
func UnZip(zipFileName, tempDestination string) string {
	zipReader, _ := zip.OpenReader(zipFileName)

	var extractedFilePath = ""
	for _, file := range zipReader.Reader.File {

		zippedFile, err := file.Open()
		errors.CheckErr(err, 402, "")
		defer zippedFile.Close()

		targetDir := tempDestination
		extractedFilePath = filepath.Join(
			targetDir,
			file.Name,
		)

		if file.FileInfo().IsDir() {
			log.Println("Directory Created:", extractedFilePath)
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			log.Println("File extracted:", file.Name)

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
	log.Println("File extracted:", zipFileName)
	fmt.Println("Extraced file path", extractedFilePath)
	return extractedFilePath
}

//MoveFiles to directory specified in command
func MoveFiles(source, destination string) {
	src := source
	dest := destination

	fmt.Println("==> moving files from ", src)
	fmt.Println("==> moving files too ", dest)

	err := os.Rename(src, dest)
	if err != nil {
		log.Fatal(err)
	}
}
