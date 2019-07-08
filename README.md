# Codewind Installer
Install Codewind on MacOS or Windows.
Prebuilt binaries are available for download [on Eclipse](https://download.eclipse.org/codewind/codewind-installer/).

[![License](https://img.shields.io/badge/License-EPL%202.0-red.svg?label=license&logo=eclipse)](https://www.eclipse.org/legal/epl-2.0/)

## Downloading the release binary for MacOS
1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the Terminal window.
3. If the binary file has the extension `.dms`, remove the extension so that the file is named `mac-installer`.
4. Enter the `chmod +x mac-installer` command to give yourself execution permissions for the binary.
5. Log in to Docker to set the environment variables.
6. If you already have a `codewind-workspace` with your projects in it, copy it into your `/Users/<username>` home directory. If you do not already have a workspace, the installer creates an empty workspace for you in this directory.
7. Type `./mac-installer` in the Terminal window with the exported environment variables to run the installer.
8. To run a command, enter `./mac-installer <command>`.

## Downloading the release binary for Linux
1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the Terminal window.
3. If necessary, remove any file extensions so that the file is named `linux-installer`.
4. Enter the `chmod +x linux-installer` command to give yourself execution permissions for the binary. 
5. Log in to Docker to set the environment variables.
6. If you already have a `codewind-workspace` with your projects in it, copy the workspace into your `$HOME` home directory. If you do not already have a workspace, the installer creates an empty workspace for you in this directory.
7. Install `docker-compose` with [Install Docker Compose](https://docs.docker.com/compose/install/).
8. To run the installer, enter `./linux-installer` in the Terminal window with the exported environment variables.
9. To run a command, enter `./linux-installer <command>`.

## Downloading the release binary for Windows
1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the command prompt.
3. Ensure that the binary file has an `.exe` extension. If it doesn't, add the extension to the file name.
4. Log in to Docker to set the environment variables.
5. Ensure that your `C:\` drive is shared with Docker Desktop for Windows. To check, go to **Settings**>**Shared drives** and make sure the `C:\` drive check box is selected.
6. If you already have a `codewind-workspace` with your projects in it, copy the `codewind-workspace` into your `C:\` directory. If you do not already have a workspace, the installer creates an empty one for you in this directory.
7. To get started and see the commands that are available, type the ` .\win-installer.exe` command in the command prompt with the exported environment variables.
8. To run a command, enter ` .\win-installer.exe <command>`.

## Build and deploying locally on MacOS
1. Ensure that you have a Go environment set up. If you don't yet have a Go environment, see [Install Go](https://golang.org/doc/install).
2. If you have Brew, use the following commands to install `dep` for MacOS:
```
$ brew install dep
$ brew upgrade dep
```
3. Clone the `git clone https://github.com/eclipse/codewind-installer.git` repo.
4. Use the `cd` command to go into the project directory and install the vendor packages with the `dep ensure -v` command.
5. Build the binary and give it a name with the `go build -o <binary-name>` command. To build a binary without the debug symbols use the command `go build -ldflags="-s -w" -o <binary-name>`.
6. Log in to Docker to set the environment variables.
7. Copy your codewind-workspace into your `/Users/<username>` home directory.
8. Type `./<binary-name>` in the Terminal window with the exported environment varibles to run the installer.
9. To run a command, enter `./<binary-name> <command>`.

## Creating a cross-platform binary
1. Use the `go tool dist list` command to get a list of the possible `GOOS/ARCH` combinations available to build.
2. Choose the `GOOS/ARCH` that you want to build for and then enter `GOOS=<OS> GOARCH=<ARCH> go build` to create the binary. To build a binary without the debug symbols use the command `GOOS=<OS> GOARCH=<ARCH> go build -ldflags="-s -w"`.

## Unit testing
1. Clone this repository.
2. Use the `cd` command to go to the test directory. The `utils.go` tests are located in the `utils/utils_test.go` file.
3. To run the tests, type the `go test -v` command in the terminal window and wait for the tests to finish.
4. For any other unit tests, the same steps apply, but the directory might change.

## Contributing
Submit issues and contributions:
1. [Submitting issues](https://github.com/eclipse/codewind/issues)
2. [Contributing](CONTRIBUTING.md)
