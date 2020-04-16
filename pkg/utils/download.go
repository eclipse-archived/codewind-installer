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
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// DownloadFromURLThenExtract downloads files from a URL
// to a destination, extracting them if necessary
func DownloadFromURLThenExtract(URL string, destination string) error {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return err
	}

	if IsTarGzURL(URL) {
		return DownloadFromTarGzURL(URL, destination)
	}
	return DownloadFromRepoURL(URL, destination)
}

// DownloadFromTarGzURL downloads a tar.gz file from a URL
// and extracts it to a destination
func DownloadFromTarGzURL(URL string, destination string) error {
	time := time.Now().Format(time.RFC3339)
	time = strings.Replace(time, ":", "-", -1) // ":" is illegal char in windows
	pathToTempFile := path.Join(os.TempDir(), "_"+time+"temp.tar.gz")

	err := DownloadFile(URL, pathToTempFile)
	if err != nil {
		return err
	}
	err = UnTar(pathToTempFile, destination)
	os.Remove(pathToTempFile)
	return err
}

// DownloadFromRepoURL downloads a repo from a URL to a destination
func DownloadFromRepoURL(repoURL string, destination string) error {
	// expecting string in format 'https://github.com/<owner>/<repo>'
	if strings.HasPrefix(repoURL, "https://") {
		repoURL = strings.TrimPrefix(repoURL, "https://")
	}
	repoArray := strings.Split(repoURL, "/")
	owner := repoArray[1]
	repo := repoArray[2]
	branch := "master"

	zipURL, err := GetZipURL(owner, repo, branch)
	if err != nil {
		return err
	}

	return DownloadAndExtractZip(zipURL, destination)
}

// DownloadAndExtractZip downloads a zip file from a URL
// and extracts it to a destination
func DownloadAndExtractZip(zipURL string, destination string) error {
	time := time.Now().Format(time.RFC3339)
	time = strings.Replace(time, ":", "-", -1) // ":" is illegal char in windows
	pathToTempZipFile := path.Join(os.TempDir(), "_"+time+".zip")

	err := DownloadFile(zipURL, pathToTempZipFile)
	if err != nil {
		return err
	}

	err = UnZip(pathToTempZipFile, destination)
	if err != nil {
		return err
	}

	os.Remove(pathToTempZipFile)
	return nil
}

// IsTarGzURL returns whether the provided URL is a tar.gz file
func IsTarGzURL(URL string) bool {
	return strings.HasSuffix(URL, ".tar.gz")
}
