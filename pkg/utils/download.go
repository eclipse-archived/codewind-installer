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

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

type (
	// GitCredentials : credentials to access GitHub or GitHubEnterprise
	GitCredentials struct {
		Username            string `json:"username,omitempty"`
		Password            string `json:"password,omitempty"`
		PersonalAccessToken string `json:"personalAccessToken,omitempty"`
	}
)

// DownloadFromURLThenExtract downloads files from a URL
// to a destination, extracting them if necessary
func DownloadFromURLThenExtract(inURL, destination string, gitCredentials *GitCredentials) error {
	URL, err := url.ParseRequestURI(inURL)
	if err != nil {
		return err
	}
	if !URL.IsAbs() {
		return fmt.Errorf("URL must be absolute, but received relative URL %s", URL)
	}

	if IsTarGzURL(URL) {
		return DownloadFromTarGzURL(URL, destination, gitCredentials)
	}
	return DownloadFromRepoURL(URL, destination, gitCredentials)
}

// DownloadFromTarGzURL downloads a tar.gz file from a URL
// and extracts it to a destination
func DownloadFromTarGzURL(URL *url.URL, destination string, gitCredentials *GitCredentials) error {
	time := time.Now().Format(time.RFC3339)
	time = strings.Replace(time, ":", "-", -1) // ":" is illegal char in windows
	pathToTempFile := path.Join(os.TempDir(), "_"+time+"temp.tar.gz")

	if gitCredentials != nil {
		downloadURL, err := getURLToDownloadReleaseAsset(URL, gitCredentials)
		if err != nil {
			return err
		}
		URL = downloadURL
	}

	err := DownloadFile(URL, pathToTempFile)
	if err != nil {
		return err
	}
	err = UnTar(pathToTempFile, destination)
	os.Remove(pathToTempFile)
	return err
}

func getURLToDownloadReleaseAsset(URL *url.URL, gitCredentials *GitCredentials) (*url.URL, error) {
	URLPathSlice := strings.Split(URL.Path, "/")

	var httpClient = &http.Client{}

	if !strings.Contains(URL.Host, "github") || len(URLPathSlice) < 6 {
		return nil, fmt.Errorf("URL must point to a GitHub repository release asset: %v", URL)
	}
	client, err := getGitHubClient(URL.Host, gitCredentials)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	owner := URLPathSlice[1]
	repo := URLPathSlice[2]
	releases, gitHubResponse, err := client.Repositories.ListReleases(
		ctx,
		owner,
		repo,
		nil,
	)
	resp := gitHubResponse.Response
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[client.Repositories.ListReleases] GitHub responded with status code %d", resp.StatusCode)
	}

	release := URLPathSlice[5]
	assetID, err := findAssetID(releases, release, URL)
	if err != nil {
		return nil, err
	}

	_, redirectURL, err := client.Repositories.DownloadReleaseAsset(
		ctx,
		owner,
		repo,
		assetID,
		httpClient,
	)
	if err != nil {
		return nil, fmt.Errorf("Repositories.DownloadReleaseAsset returned error: %v", err)
	}
	ru, _ := url.ParseRequestURI(redirectURL)
	return ru, nil
}

func findAssetID(releases []*github.RepositoryRelease, releaseName string, URL *url.URL) (int64, error) {
	for _, v := range releases {
		if *v.TagName == releaseName {
			for _, val := range v.Assets {
				if *val.BrowserDownloadURL == URL.String() {
					return *val.ID, nil
				}
			}
			return 0, fmt.Errorf("Cannot find matching assets for release %s", releaseName)
		}
	}
	return 0, fmt.Errorf("Cannot find release %s", releaseName)
}

// DownloadFromRepoURL downloads a repo from a URL to a destination
func DownloadFromRepoURL(URL *url.URL, destination string, gitCredentials *GitCredentials) error {
	URLPathSlice := strings.Split(URL.Path, "/")

	if !strings.Contains(URL.Host, "github") || len(URLPathSlice) < 3 {
		return fmt.Errorf("URL must point to a GitHub repository release asset: %v", URL)
	}

	client, err := getGitHubClient(URL.Host, gitCredentials)
	if err != nil {
		return err
	}

	owner := URLPathSlice[1]
	repo := URLPathSlice[2]
	zipURL, err := GetZipURL(owner, repo, "master", client)
	if err != nil {
		return err
	}

	return DownloadAndExtractZip(zipURL, destination)
}

func getGitHubClient(domain string, gitCredentials *GitCredentials) (*github.Client, error) {
	if gitCredentials == nil {
		return github.NewClient(nil), nil
	}

	if gitCredentials.PersonalAccessToken != "" {
		ctx := context.Background()
		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: gitCredentials.PersonalAccessToken},
		)
		tokenClient := oauth2.NewClient(ctx, tokenSource)
		baseURL := "https://" + domain
		return github.NewEnterpriseClient(baseURL, baseURL, tokenClient)
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
func GetZipURL(owner, repo, branch string, client *github.Client) (*url.URL, error) {
	ctx := context.Background()
	opt := &github.RepositoryContentGetOptions{Ref: branch}
	URL, _, err := client.Repositories.GetArchiveLink(ctx, owner, repo, github.Zipball, opt, true)
	if err != nil {
		return nil, err
	}
	return URL, nil
}

// DownloadAndExtractZip downloads a zip file from a URL
// and extracts it to a destination
func DownloadAndExtractZip(zipURL *url.URL, destination string) error {
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

// DownloadFile from URL to file destination
func DownloadFile(URL *url.URL, destination string) error {
	resp, err := http.Get(URL.String())
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
func IsTarGzURL(URL *url.URL) bool {
	return strings.HasSuffix(URL.Path, ".tar.gz")
}
