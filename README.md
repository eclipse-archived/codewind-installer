# Codewind Installer
Install Codewind on MacOS or Windows.
Prebuilt binary files are available for download [on Eclipse](https://download.eclipse.org/codewind/codewind-installer/).

[![License](https://img.shields.io/badge/License-EPL%202.0-red.svg?label=license&logo=eclipse)](https://www.eclipse.org/legal/epl-2.0/)
[![Build Status](https://ci.eclipse.org/codewind/buildStatus/icon?job=Codewind%2Fcodewind-installer%2Fmaster)](https://ci.eclipse.org/codewind/job/Codewind/job/codewind-installer/job/master/)

## Before starting

Ensure that you are logged in to Docker. Type `docker login` into a command line window and follow the instructions.

## Downloading the release binary file for MacOS

1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the command line window.
3. If the binary file has the `.dms` extension, remove the extension so that the file is named `codewind-installer-macos`.
4. Enter the `chmod +x codewind-installer-macos` command to give yourself execution permissions for the binary.
5. If you already have a `codewind-workspace` with your projects in it, copy it into your `/Users/<username>` home directory. If you do not already have a workspace, the installer creates an empty workspace for you in this directory.
6. Type `./codewind-installer-macos` in the command line window to run the installer.
7. To run a command, enter `./codewind-installer-macos <command>`.

## Downloading the release binary for Linux

1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the command line window.
3. If necessary, remove any file extensions so that the file is named `codewind-installer-linux`.
4. Enter the `chmod +x codewind-installer-linux` command to give yourself execution permissions for the binary.
5. If you already have a `codewind-workspace` with your projects in it, copy the workspace into your `$HOME` home directory. If you do not already have a workspace, the installer creates an empty workspace for you in this directory.
6. Install `docker-compose` with [Install Docker Compose](https://docs.docker.com/compose/install/).
7. To run the installer, enter `./codewind-installer-linux` in the command line window.
8. To run a command, enter `./codewind-installer-linux <command>`.

## Downloading the release binary for Windows

1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the command prompt.
3. Ensure that the binary file has an `.exe` extension. If it doesn't, add the extension to the file name.
4. Ensure that your `C:\` drive is shared with Docker Desktop for Windows. To check, go to **Settings**>**Shared drives** and make sure the `C:\` drive check box is selected.
5. If you already have a `codewind-workspace` with your projects in it, copy the `codewind-workspace` into your `C:\` directory. If you do not already have a workspace, the installer creates an empty one for you in this directory.
6. To get started and see the commands that are available, type the `.\codewind-installer-win.exe` command in the command prompt.
7. To run a command, enter `.\codewind-installer-win.exe <command>`.

## Building and deploying locally on MacOS

1. Ensure that you have a Go environment set up. If you don't yet have a Go environment, see [Install Go](https://golang.org/doc/install).
2. If you have Brew, use the following commands to install `dep` for MacOS:

```
$ brew install dep
$ brew upgrade dep
```

3. Create the directory tree expected `../go/src/github.com/eclipse`.
4. Use the `cd` command to go to the `eclipse` directory that you previously created.
5. After you go to this directory, clone the repository by typing `git clone https://github.com/eclipse/codewind-installer.git`.
6. Use the `cd` command to go to the project directory and install the vendor packages with the `dep ensure -v` command.
7. Build the binary and give it a name with the `go build -o <binary-name>` command. To build a binary without the debug symbols, use the `go build -ldflags="-s -w" -o <binary-name>` command.
8. Copy your `codewind-workspace` into your `/Users/<username>` home directory.
9. Type `./<binary-name>` in the command line window to run the installer.
10. To run a command, enter `./<binary-name> <command>`.

## Creating a cross-platform binary

1. Use the `go tool dist list` command to get a list of the possible `GOOS/ARCH` combinations available to build.
2. Choose the `GOOS/ARCH` that you want to build for and then enter `GOOS=<OS> GOARCH=<ARCH> go build` to create the binary. To build a binary without the debug symbols, use the `GOOS=<OS> GOARCH=<ARCH> go build -ldflags="-s -w"` command.

## Unit testing

1. Clone the `codewind-installer` repository.
2. Use the `cd` command to go to a directory with test files in. For example, the `utils.go` tests are located in the `utils/utils_test.go` file.
3. To run the tests, enter the `go test -v` command in the command line window and wait for the tests to finish.
4. For any other unit tests, the same steps apply, but the directory might change.

## Bats-core testing

1. Set up your environment by installing bats-core as per the bats-core instructions found at <https://github.com/bats-core/bats-core>.
2. Clone the `codewind-installer` repository.
3. Use the `cd` command to go to the top level project directory.
4. Ensure your system environment is clean by having no Codewind images installed or containers running.
5. To run the tests, enter the `bats integration.bats` command in the command line window and wait for the tests to finish.

## CLI Commands

|Command         |Alias         |Usage                                                               |
|----------------|--------------|--------------------------------------------------------------------|
|project         |              |'Manage Codewind projects'                                          |
|install         |`in`          |'Pull pfe, performance & initialize images from dockerhub'          |
|start           |              |'Start the Codewind containers'                                     |
|status          |              |'Print the installation status of Codewind'                         |
|stop            |              |'Stop the running Codewind containers'                              |
|stop-all        |              |'Stop all of the Codewind and project containers'                   |
|remove          |`rm`          |'Remove Codewind/Project docker images and the codewind network'    |
|templates       |              |'Manage project templates'                                          |
|sectoken        |`st`          |'Authenticate with username and password to obtain an access_token' |
|secrealm        |`sr`          |'Manage new or existing REALM configurations'                       |
|secclient       |`sc`          |'Manage new or existing APPLICATION access configurations'          |
|secuser         |`su`          |'Manage new or existing USER access configurations'                 |
|help            |`h`           |'Shows a list of commands or help for one command'                  |

## CLI Command Options

## project

`--url/-u <value>` - URL of project to download

## install

`--tag/-t <value>` - Dockerhub image tag (default: "latest")</br>
`--json/-j` - Specify terminal output

## start

`--tag/-t <value>` - Dockerhub image tag (default: "latest")</br>
`--debug/-d` - Add debug output

## status

`--json/-j` - Specify terminal output

## stop

>**Note:** No additional flags

## stop-all

>**Note:** No additional flags

## remove

`--tag/-t <value>` - Dockerhub image tag.</br>
**Note:** Failing to specify a `--tag`, will remove all Codewind images on the host machine.

## templates

>**Note:** No additional flags

Subcommands:</br>

`list/ls` - List available templates

## sectoken

Subcommands:</br>

`get/g` - Authenticate and obtain an access_token

> **Flags:**
> --host value                  URL or ingress to Keycloak service
> --realm value                 Application realm
> --username value              Account Username
> --password value              Account Password
> --client value                Client

## secrealm

Subcommands:</br>

`create/c` - Create a new realm (requires either admin_token or username/password)

> **Flags:**
> --host value                   URL or ingress to Keycloak service
> --realm value                  Application realm
> --accesstoken value            Admin access_token
> --username value               Admin Username
> --password value               Admin Password

## secclient

Subcommands:</br>

`create/c` - Create a new client in an existing Keycloak realm (requires either admin_token or username/password)

> --host value                   URL or ingress to Keycloak service
> --realm value                  Application realm
> --clientid value               New client ID to create
> --redirect value               Allowed redirect callback URL eg: `http://127.0.0.1:9090/*`
> --accesstoken value            Admin access_token
> --username value               Admin Username
> --password value               Admin Password

`get/g` - Get client id (requires either admin_token or username/password)

> --host value                   URL or ingress to Keycloak service
> --realm value                  Application realm
> --clientid value               New client ID to create
> --accesstoken value            Admin access_token
> --username value               Admin Username
> --password value               Admin Password

`secret/s` - Get client secret (requires either admin_token or username/password)

> --host value                   URL or ingress to Keycloak service
> --realm value                  Application realm
> --clientid value               New client ID to create
> --accesstoken value            Admin access_token
> --username value               Admin Username
> --password value               Admin Password

## secuser

Subcommands:</br>

`create/c` - Create a new user in an existing Keycloak realm (requires either admin_token or username/password)

> --host value                   URL or ingress to Keycloak service
> --realm value                  Application realm
> --accesstoken value            Admin access_token
> --username value               Admin Username
> --password value               Admin Password
> --name value                   Username to add

`get/g` - Gets an existing Keycloak user from an existing realm (requires either admin_token or username/password)

> --host value                   URL or ingress to Keycloak service
> --realm value                  Application realm
> --accesstoken value            Admin access_token
> --username value               Admin Username
> --password value               Admin Password
> --name value                   Username to query

`setpw/p` - Reset an existing users password (requires either admin_token or username/password)

> --host value                   URL or ingress to Keycloak service
> --realm value                  Application realm
> --accesstoken value            Admin access_token
> --username value               Admin Username
> --password value               Admin Password
> --name value                   Username to query
> --newpw value                  New replacement password

## help

`--help/-h` - Shows a list of commands or help for one command

## Contributing

Submit issues and contributions:

1. [Submitting issues](https://github.com/eclipse/codewind/issues)
2. [Contributing](CONTRIBUTING.md)
