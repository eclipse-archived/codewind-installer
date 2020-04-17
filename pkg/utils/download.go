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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

type (
	// GitCredentials : credentials to access GitHub or GitHubEnterprise
	GitCredentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
)

// DownloadFromURLThenExtract downloads files from a URL
// to a destination, extracting them if necessary
func DownloadFromURLThenExtract(URL, destination string, gitCredentials GitCredentials) error {
	_url, err := url.ParseRequestURI(URL)
	if err != nil {
		return err
	}
	if !_url.IsAbs() {
		return fmt.Errorf("URL must be absolute, but received relative URL %s", _url)
	}

	if IsTarGzURL(URL) {
		return DownloadFromTarGzURL(URL, destination, gitCredentials)
	}
	return DownloadFromRepoURL(URL, destination, gitCredentials)
}

// DownloadFromTarGzURL downloads a tar.gz file from a URL
// and extracts it to a destination
func DownloadFromTarGzURL(URL, destination string, gitCredentials GitCredentials) error {
	time := time.Now().Format(time.RFC3339)
	time = strings.Replace(time, ":", "-", -1) // ":" is illegal char in windows
	pathToTempFile := path.Join(os.TempDir(), "_"+time+"temp.tar.gz")

	if gitCredentials != (GitCredentials{}) {
		downloadURL, err := getURLToDownloadReleaseAsset(URL, gitCredentials)
		if err != nil {
			return err
		}
		URL = downloadURL
	}

	err := DownloadFile(URL, pathToTempFile, gitCredentials)
	if err != nil {
		return err
	}
	err = UnTar(pathToTempFile, destination)
	os.Remove(pathToTempFile)
	return err
}

func getURLToDownloadReleaseAsset(URL string, gitCredentials GitCredentials) (string, error) {
	URLSlice := strings.Split(URL, "/")

	domain := URLSlice[2]
	client, err := getGitHubClient(domain, gitCredentials)
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	owner := URLSlice[3]
	repo := URLSlice[4]
	releases, gitHubResponse, err := client.Repositories.ListReleases(
		ctx,
		owner,
		repo,
		nil,
	)
	resp := gitHubResponse.Response
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("[client.Repositories.ListReleases] GitHub responded with status code %d", resp.StatusCode)
	}

	release := URLSlice[7]
	assetID, err := findAssetID(releases, release, URL)
	if err != nil {
		return "", err
	}

	_, redirectURL, err := client.Repositories.DownloadReleaseAsset(
		ctx,
		owner,
		repo,
		assetID,
	)
	if err != nil {
		return "", fmt.Errorf("Repositories.DownloadReleaseAsset returned error: %v", err)
	}
	return redirectURL, nil
}

func findAssetID(releases []*github.RepositoryRelease, releaseName, URL string) (int64, error) {
	for _, v := range releases {
		if *v.TagName == releaseName {
			for _, val := range v.Assets {
				if *val.BrowserDownloadURL == URL {
					return *val.ID, nil
				}
			}
			return 0, fmt.Errorf("Cannot find matching assets for release %s", releaseName)
		}
	}
	return 0, fmt.Errorf("Cannot find release %s", releaseName)
}

// DownloadFromRepoURL downloads a repo from a URL to a destination
func DownloadFromRepoURL(repoURL, destination string, gitCredentials GitCredentials) error {
	// expecting string in format 'https://<github.com>/<owner>/<repo>'
	URLSlice := strings.Split(repoURL, "/")

	domain := URLSlice[2]
	client, err := getGitHubClient(domain, gitCredentials)
	if err != nil {
		return err
	}

	owner := URLSlice[3]
	repo := URLSlice[4]
	zipURL, err := GetZipURL(owner, repo, "master", client)
	if err != nil {
		return err
	}

	return DownloadAndExtractZip(zipURL, destination)
}

func getGitHubClient(domain string, gitCredentials GitCredentials) (*github.Client, error) {
	if gitCredentials == (GitCredentials{}) {
		return github.NewClient(nil), nil
	}

	tp := github.BasicAuthTransport{
		Username: gitCredentials.Username,
		Password: gitCredentials.Password,
	}
	if domain == "github.com" {
		return github.NewClient(tp.Client()), nil
	}
	baseURL := "https://" + domain
	return github.NewEnterpriseClient(baseURL, baseURL, tp.Client())
}

// GetZipURL from github api /repos/:owner/:repo/:archive_format/:ref
func GetZipURL(owner, repo, branch string, client *github.Client) (string, error) {
	ctx := context.Background()
	opt := &github.RepositoryContentGetOptions{Ref: branch}
	URL, _, err := client.Repositories.GetArchiveLink(ctx, owner, repo, "zipball", opt, true)
	if err != nil {
		return "", err
	}
	url := URL.String()
	return url, nil
}

// DownloadAndExtractZip downloads a zip file from a URL
// and extracts it to a destination
func DownloadAndExtractZip(zipURL string, destination string) error {
	time := time.Now().Format(time.RFC3339)
	time = strings.Replace(time, ":", "-", -1) // ":" is illegal char in windows
	pathToTempZipFile := path.Join(os.TempDir(), "_"+time+".zip")

	err := DownloadFile(zipURL, pathToTempZipFile, GitCredentials{})
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

// DownloadFile from URL to file destination
func DownloadFile(URL, destination string, gitCredentials GitCredentials) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("File download failed for %s, status code %d", URL, resp.StatusCode)
	}

	// Create the file
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write body to file
	_, err = io.Copy(file, resp.Body)
	return err
}

// IsTarGzURL returns whether the provided URL is a tar.gz file
func IsTarGzURL(URL string) bool {
	return strings.HasSuffix(URL, ".tar.gz")
}
