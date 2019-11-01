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
	"errors"
	"fmt"
	"net/http"
	"time"
)

// HTTPClient : An net HTTP Client to simplify testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// WaitForService : Wait for service to start
func WaitForService(url string, successStatusCode int, maxRetries int) error {
	retries := 0
	client := http.Client{
		Timeout: time.Second * 5,
	}
	for {
		response, err := client.Get(url)
		if err == nil && response.StatusCode == successStatusCode {
			fmt.Println(".")
			return nil
		}
		fmt.Print(".")
		time.Sleep(1 * time.Second)
		retries++
		if retries == maxRetries {
			break
		}
	}
	fmt.Println(".")
	return errors.New("Service did not respond")
}
