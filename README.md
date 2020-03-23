# Codewind Command Line Interface (CLI)

Install Codewind on MacOS or Windows.
Prebuilt binary files are available for download [on Eclipse](https://download.eclipse.org/codewind/codewind-installer/).

[![License](https://img.shields.io/badge/License-EPL%202.0-red.svg?label=license&logo=eclipse)](https://www.eclipse.org/legal/epl-2.0/)
[![Build Status](https://ci.eclipse.org/codewind/buildStatus/icon?job=Codewind%2Fcodewind-installer%2Fmaster)](https://ci.eclipse.org/codewind/job/Codewind/job/codewind-installer/job/master/)
[![Chat](https://img.shields.io/static/v1.svg?label=chat&message=mattermost&color=145dbf)](https://mattermost.eclipse.org/eclipse/channels/eclipse-codewind)
[![Go Report Card](https://goreportcard.com/badge/github.com/eclipse/codewind-installer)](https://goreportcard.com/report/github.com/eclipse/codewind-installer)
[![codecov](https://codecov.io/gh/eclipse/codewind-installer/branch/master/graph/badge.svg)](https://codecov.io/gh/eclipse/codewind-installer)

## Table of Contents

- [Before starting](#before-starting)
- [Downloading the release binary file for](#Downloading-the-release-binary-file-for)
- [Building and deploying locally on MacOS](#Building-and-deploying-locally-on-MacOS)
- [Creating a cross-platform binary](#Creating-a-cross-platform-binary)
- [Running the Tests](#running-the-tests)
- [API](#API)
- [Contributing](#Contributing)

## Before starting

Ensure that you are logged in to Docker. Type `docker login` into a command line window and follow the instructions.

## Downloading the release binary file for:

### MacOS

1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the command line window.
3. If the binary file has the `.dms` extension, remove the extension so that the file is named `cwctl-macos`.
4. Enter the `chmod +x cwctl-macos` command to give yourself execution permissions for the binary.
5. If you already have a `codewind-workspace` with your projects in it, copy it into your `/Users/<username>` home directory. If you do not already have a workspace, the CLI creates an empty workspace for you in this directory.
6. Type `./cwctl-macos` in the command line window to run the CLI.
7. To run a command, enter `./cwctl-macos <command>`.

### Linux

1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the command line window.
3. If necessary, remove any file extensions so that the file is named `cwctl-linux`.
4. Enter the `chmod +x cwctl-linux` command to give yourself execution permissions for the binary.
5. If you already have a `codewind-workspace` with your projects in it, copy the workspace into your `$HOME` home directory. If you do not already have a workspace, the CLI creates an empty workspace for you in this directory.
6. Install `docker-compose` with [Install Docker Compose](https://docs.docker.com/compose/install/).
7. To run the CLI, enter `./cwctl-linux` in the command line window.
8. To run a command, enter `./cwctl-linux <command>`.

### Windows

1. Download the release binary file to a folder on your system.
2. Use the `cd` command to go to the location of the downloaded file in the command prompt.
3. Ensure that the binary file has an `.exe` extension. If it doesn't, add the extension to the file name.
4. Ensure that your `C:\` drive is shared with Docker Desktop for Windows. To check, go to **Settings**>**Shared drives** and make sure the `C:\` drive check box is selected.
5. If you already have a `codewind-workspace` with your projects in it, copy the `codewind-workspace` into your `C:\` directory. If you do not already have a workspace, the CLI creates an empty one for you in this directory.
6. To get started and see the commands that are available, type the `.\cwctl-win.exe` command in the command prompt.
7. To run a command, enter `.\cwctl-win.exe <command>`.

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
7. Using the `cd` command, navigate into `/cmd/cli/` within the project directory.
8. Build the binary and give it a name with the `go build -o <binary-name>` command. To build a binary without the debug symbols, use the `go build -ldflags="-s -w" -o <binary-name>` command.
9. Copy your `codewind-workspace` into your `/Users/<username>` home directory.
10. Type `./<binary-name>` in the command line window to run the CLI.
11. To run a command, enter `./<binary-name> <command>`.

## Creating a cross-platform binary

1. Use the `go tool dist list` command to get a list of the possible `GOOS/ARCH` combinations available to build.
2. Choose the `GOOS/ARCH` that you want to build for and then enter `GOOS=<OS> GOARCH=<ARCH> go build` to create the binary. To build a binary without the debug symbols, use the `GOOS=<OS> GOARCH=<ARCH> go build -ldflags="-s -w"` command.

## Running the Tests

### Integration tests (Bats-core tests)

1. Set up your environment by installing bats-core as per the bats-core instructions found at <https://github.com/bats-core/bats-core>.
2. Clone the `codewind-installer` repository.
3. Use the `cd` command to go to the top level project directory.
4. Ensure your system environment is clean by having no Codewind images installed or containers running.
5. To run the tests, enter the `bats integration.bats` command in the command line window and wait for the tests to finish.

### Unit tests

#### All tests in current directory and all of its subdirectories

```bash
$ go test ./...
```

#### All tests from a package (e.g. utils)

```bash
$ go test github.com/eclipse/codewind-installer/utils
```

#### All tests from a package (e.g. utils) whose names match a regex

```bash
$ go test github.com/eclipse/codewind-installer/utils -run ^(TestDetermineProjectInfo)$ # or ^(DetermineProjectInfo)$
```

```bash
$ go test github.com/eclipse/codewind-installer/utils -run ^(TestDetermineProjectInfo)$
```

#### Notes

##### In VSCode, the UI and Command Palette provide easy ways to run specific packages and tests

##### If tests pass, `go test` will just report that tests passed

E.g.:

```
PASS
ok  	github.com/eclipse/codewind-installer/utils
```

To see more details, use `go test -v`.

### Test coverage

```bash
$ ./test.sh -coverage
```

This is the same as `go test ./...`, but outputs test coverage scores for each function, package, and an overall score.

## API

### CLI Commands

| Command         | Alias | Usage                                                               |
| ----------------| ----- | ------------------------------------------------------------------- |
| project         |       | 'Manage Codewind projects'                                          |
| install         | `in`  | 'Pull pfe & performance images from dockerhub'                      |
| start           |       | 'Start the Codewind containers'                                     |
| status          |       | 'Print the installation status of Codewind'                         |
| stop            |       | 'Stop the running Codewind containers'                              |
| stop-all        |       | 'Stop all of the Codewind and project containers'                   |
| remove          | `rm`  | 'Remove Codewind and Project docker images'                         |
| templates       |       | 'Manage project templates'                                          |
| version         |       | 'Print the versions of Codewind containers, for a given connection' |
| sectoken        | `st`  | 'Authenticate with username and password to obtain an access_token' |
| secrole         | `sl`  | 'Manage realm based ACCESS roles'                                   |
| secrealm        | `sr`  | 'Manage new or existing REALM configurations'                       |
| secclient       | `sc`  | 'Manage new or existing APPLICATION access configurations'          |
| seckeyring      | `sk`  | 'Manage Codewind keys in the desktop keyring'                       |
| secuser         | `su`  | 'Manage new or existing USER access configurations'                 |
| connections     | `con` | 'Manage connections configuration list'                             |
| loglevels       | `log` | 'Get or set logging levels for Codewind containers'                 |
| registrysecrets | `rs`  | 'Manage docker registry secrets'                                    |
| help            | `h`   | 'Shows a list of commands or help for one command'                  |

### Command Options:

### project

`--url/-u <value>` - URL of project to download

Subcommands:</br>

`create` - Downloads a project created from a template, at the given URL

> **Flags:**
> --url,-u value URL of project to download
> --path,-p value Path at which to create the new project
> --conid value Connection ID of PFE that will be used to validate the project (optional)

`validate` - Returns the predicted language and build type for a project, and writes a default .cw-settings to it if one does not already exist

> **Flags:**
> --path,-p value Project path, on local disk
> --type,-t value Project build type, if known (not required)
> --conid value Connection ID of PFE that will be used to validate the project (optional)

`bind` - Bind a project to Codewind for building and running

> **Flags:**
> --name,-n value Project name
> --language,-l value Project language
> --type,-t value Project Type
> --path,-p value Project Path
> --conid value Connection ID

`sync` - Synchronize a bound project to its connection

> **Flags:**
> --path,-p value Project Path
> --id,-i value Project ID
> --time,-t value Time of last project sync

`list` - List projects bound to a Codewind deployment
> **Flags**
> --conid value                 Connection ID

`get` - Get a single project, requires either the project ID or name
When using a project ID the CLI will automatically detect which connection it relates to
> **Flags**
> --id value                    Project ID
> --name                        Project name
> --conid                       Connection ID

`restart` - Restart a project
> **Flags**
> --id, i                       Project ID
> --conid                       Connection ID
> --startMode                   "run" | "debug" | "debugNoInit"

`connection/con` - Manage the connection targets for a project

`set,s` - Sets the connection for a projectID

> **Flags:**
> --id,i value Project ID
> --conid value Connection ID

`get,g` - Gets connections for a projectID

> **Flags:**
> --id,i value Project ID

`remove,r` - Removes the connection from a projectID

> **Flags:**
> --id,i value Project ID

## install

`--tag/-t <value>` - Dockerhub image tag (default: "latest")</br>
`--json/-j` - Specify terminal output

Subcommands:</br>

`remote` - Install a remote deployment of Codewind

> **Flags:**
> --namespace,-n value Kubernetes namespace to install into
> --session,-ses value Codewind session secret to encrypt session store
> --ingress,-i value Ingress Domain eg: 10.22.33.44.nip.io
> --kadminuser,-au value Keycloak admin user
> --kadminpass,-ap value Keycloak admin password
> --kdevuser,-du value Keycloak developer username
> --kdevpass,-dp value Keycloak developer username initial password
> --krealm,-r value Keycloak realm to setup
> --kclient,-c value Keycloak client to setup
> --pvcsize,-p value Codewind PVC size (integer between 1 and 999 Gigabytes)
> --kurl value Don't deploy a new Keycloak pod, use an existing one at this URL
> --konly Install a deployment of Keycloak only

### start

`--tag/-t <value>` - Dockerhub image tag (default: "latest")</br>
`--debug/-d` - Add debug output

### status

`--json/-j` - Specify terminal output

### stop

> **Note:** No additional flags

### stop-all

> **Note:** No additional flags

### remove

`--tag/-t <value>` - Dockerhub image tag.</br>
> **Note:** Failing to specify a `--tag`, will result in an attempt to remove the default `latest` tagged Codewind images on the host machine.

Subcommands:</br>

`local/l` - Removes and deletes a Codewind local deployment
> **Flags:**
> --tag - Docker hub image tag

`remote/r` - Removes and deletes a Codewind remote deployment from Kubernetes
> **Flags:**
> --namespace - Kubernetes namespace
> --workspace - Codewind workspace ID

`keycloak/k` - Removes and deletes a Keycloak deployment from Kubernetes
> **Flags:**
> --namespace - Kubernetes namespace
> --workspace - Keycloak workspace ID

### templates

> **Note:** No additional flags

Subcommands:</br>

`list/ls` - List available templates

### version

> **Flags:**
> --conid value Connection ID (see the connections cmd)
> --all - Show Container versions for all Codewind connections

## sectoken

Subcommands:</br>

`get/g` - Authenticate and obtain an access_token.

> **Note 1:**: The preferred way to authenticate is by supplying just the connection ID (conid) and username. In this mode the command will use the stored password from the platform keyring
> **Note 2:**: If you dont have a connection ID (conid) you must supply use the host, realm and client flags
> **Note 3:**: You can use a combination of both the connection ID (conid) and host/realm/client flags. In this mode, the host/realm/client flags take precedence override the connection defaults
> **Note 4:**: The password flag is optional when used with the connection ID (conid) flag and when a password already exists in the platform keyring. Including the password flag will update the keychain password after a successful login or add a password to the keychain if one does not exist

> **Flags:**
> --host value URL or ingress to Keycloak service
> --realm value Application realm
> --username value Account Username
> --password value Account Password
> --client value Client
> --conid value Connection ID (see the connections cmd)

## sectoken

Subcommands:</br>

`refresh/r` - Refresh access_token using cached refresh_token

Refresh tokens are automatically stored in the platform keychain. This command will use the refresh token to obtain a new access token from the authentication service. The access_token can then be used by curl or socket connections when accessing Codewind.

> **Flags:**
> --conid value Connection ID (see the connections cmd)

## secrealm

Subcommands:</br>

`create/c` - Create a new realm (requires either admin_token or username/password)

> **Flags:**
> --host value URL or ingress to Keycloak service
> --newrealm value Application realm to be created
> --accesstoken value Admin access_token

## secclient

Subcommands:</br>

`create/c` - Create a new client in an existing Keycloak realm (requires either admin_token or username/password)

> --host value URL or ingress to Keycloak service
> --realm value Application realm where client should be created
> --newclient value New client ID to create
> --redirect value Allowed redirect callback URL eg: `http://127.0.0.1:9090/*`
> --accesstoken value Admin access_token

`get/g` - Get client id (requires either admin_token or username/password)

> --host value URL or ingress to Keycloak service
> --realm value Application realm
> --clientid value Client ID to retrieve
> --accesstoken value Admin access_token
> --username value Admin Username
> --password value Admin Password

`secret/s` - Get client secret (requires either admin_token or username/password)

> --host value URL or ingress to Keycloak service
> --realm value Application realm
> --clientid value Client ID to retrieve
> --accesstoken value Admin access_token
> --username value Admin Username
> --password value Admin Password

## seckeyring

Subcommands:</br>

`update/u` - Add new or update existing Codewind credentials key in keyring

> --conid `<value>` Connection ID (see the connections cmd)
> --username `<value>` Username
> --password `<value>` Password

`validate/v` - Checks if credentials key exist in the keyring

> --conid `<value>` Connection ID (see the connections cmd)
> --username `<value>` Username

## secuser

Subcommands:</br>

`create/c` - Create a new user in an existing Keycloak realm (requires either admin_token or username/password)

> --host value URL or ingress to Keycloak service
> --realm value Application realm
> --accesstoken value Admin access_token
> --username value Admin Username
> --password value Admin Password
> --name value Username to add

`get/g` - Gets an existing Keycloak user from an existing realm (requires either admin_token or username/password)

> --host value URL or ingress to Keycloak service
> --realm value Application realm
> --accesstoken value Admin access_token
> --username value Admin Username
> --password value Admin Password
> --name value Username to query

`setpw/p` - Reset an existing users password (requires either admin_token or username/password)

> --host value URL or ingress to Keycloak service
> --realm value Application realm
> --accesstoken value Admin access_token
> --username value Admin Username
> --password value Admin Password
> --name value Username to query
> --newpw value New replacement password

`addrole/p` - Adds an existing role to a user (requires either admin_token or username/password)

> --host value URL or ingress to Keycloak service
> --realm value Application realm
> --accesstoken value Admin access_token
> --name value Username to target
> --role value Name of an existing role to add

## connections

Subcommands:</br>

`add/a` - Add a new connection to the list

> **Flags:**
> --label value A displayable name
> --url value The ingress URL of the PFE instance

`update/u` - Update an existing connection

> **Flags:**
> --conid value The Connection ID to update
> --label value A displayable name
> --url value The ingress URL of the PFE instance

`get/g` - Get a connection using its ID

> **Flags:**
> --conid value The Connection ID to retrieve

`remove/rm` - Remove a connection from the list

> **Flags:**
> --conid value A Connection ID

`list/ls` - List known connections

> **Note:** No additional flags

`reset` - Resets the connections list to a single local connection

> **Note:** No additional flags

## loglevels

> **Flags:**
> --conid value The Connection ID of the remote Codewind installation. Defaults to `local`.

> **Arguments:**
> The log level to set, one of `error`, `warn`, `info`, `debug`, `trace`

## remote

Subcommands:</br>

`list/l` - List all remote deployments of Codewind

> **Flags:**
> --namespace value The namespace to check (defaults to all)

## registrysecrets

Subcommands:</br>

`add/a` - Add a new docker registry secret and return the updated list of secrets

> **Flags:**
> --conid value Connection ID (see the connections cmd). Defaults to `local`.
> --address value The address of the docker registry
> --username value The username for the docker registry
> --password value The password for the docker registry

`list/ls` - List the docker secrets (registries and usernames)

> **Flags:**
> --conid value Connection ID (see the connections cmd). Defaults to `local`.

`remove/rm` - Remove a docker registry secret and return the updated list of secrets

> **Flags:**
> --conid value Connection ID (see the connections cmd). Defaults to `local`.
> --address value The address of the docker registry to remove

## help

`--help/-h` - Shows a list of commands or help for one command

## Contributing

Submit issues and contributions:

1. [Submitting issues](https://github.com/eclipse/codewind/issues)
2. [Contributing](CONTRIBUTING.md)
