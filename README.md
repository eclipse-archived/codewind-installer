# Codewind Installer
Install Codewind on MacOS or Windows.
Prebuilt binaries are available for download [on Eclipse](https://download.eclipse.org/codewind/codewind-installer/).

[![License](https://img.shields.io/badge/License-EPL%202.0-red.svg?label=license&logo=eclipse)](https://www.eclipse.org/legal/epl-2.0/)
[![Build Status](https://ci.eclipse.org/codewind/buildStatus/icon?job=Codewind%2Fcodewind-installer%2Fmaster)](https://ci.eclipse.org/codewind/job/Codewind/job/codewind-installer/job/master/)

## Before starting

Ensure you are logged into docker by typing `docker login`into any Terminal window or command prompt and follow any further instructions.

## Downloading the release binary for MacOS

1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the Terminal window.
3. If the binary file has the extension `.dms`, remove the extension so that the file is named `codewind-installer-macos`.
4. Enter the `chmod +x codewind-installer-macos` command to give yourself execution permissions for the binary.
5. If you already have a `codewind-workspace` with your projects in it, copy it into your `/Users/<username>` home directory. If you do not already have a workspace, the installer creates an empty workspace for you in this directory.
6. Type `./codewind-installer-macos` in the Terminal window to run the installer.
7. To run a command, enter `./codewind-installer-macos <command>`.

## Downloading the release binary for Linux

1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the Terminal window.
3. If necessary, remove any file extensions so that the file is named `codewind-installer-linux`.
4. Enter the `chmod +x codewind-installer-linux` command to give yourself execution permissions for the binary.
5. If you already have a `codewind-workspace` with your projects in it, copy the workspace into your `$HOME` home directory. If you do not already have a workspace, the installer creates an empty workspace for you in this directory.
6. Install `docker-compose` with [Install Docker Compose](https://docs.docker.com/compose/install/).
7. To run the installer, enter `./codewind-installer-linux` in the Terminal window.
8. To run a command, enter `./codewind-installer-linux <command>`.

## Downloading the release binary for Windows

1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the command prompt.
3. Ensure that the binary file has an `.exe` extension. If it doesn't, add the extension to the file name.
4. Ensure that your `C:\` drive is shared with Docker Desktop for Windows. To check, go to **Settings**>**Shared drives** and make sure the `C:\` drive check box is selected.
5. If you already have a `codewind-workspace` with your projects in it, copy the `codewind-workspace` into your `C:\` directory. If you do not already have a workspace, the installer creates an empty one for you in this directory.
6. To get started and see the commands that are available, type the `.\codewind-installer-win.exe` command in the command promp.
7. To run a command, enter `.\codewind-installer-win.exe <command>`.

## Build and deploying locally on MacOS

1. Ensure that you have a Go environment set up. If you don't yet have a Go environment, see [Install Go](https://golang.org/doc/install).
2. If you have Brew, use the following commands to install `dep` for MacOS:

```
$ brew install dep
$ brew upgrade dep
```

3. Create the directory tree expected `../go/src/github.com/eclipse`.
4. Use the `cd` command to navigate into the `eclipse` directory created above.
5. Once in this directory, clone the repository by typing `git clone https://github.com/eclipse/codewind-installer.git`.
6. Use the `cd` command to go into the project directory and install the vendor packages with the `dep ensure -v` command.
7. Build the binary and give it a name with the `go build -o <binary-name>` command. To build a binary without the debug symbols use the command `go build -ldflags="-s -w" -o <binary-name>`.
8. Copy your codewind-workspace into your `/Users/<username>` home directory.
9. Type `./<binary-name>` in the Terminal window to run the installer.
10. To run a command, enter `./<binary-name> <command>`.

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
