#!groovy​

pipeline {

    agent {
        kubernetes {
              label 'go-pod'
            yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: go
    image: golang:1.12-stretch
    tty: true
    command:
    - cat
    resources:
      limits:
        memory: "4Gi"
        cpu: "1"
      requests:
        memory: "4Gi"
        cpu: "1"
"""
        }
    }

    options {
        timestamps()
        skipStagesAfterUnstable()
    }

    environment {
        CODE_DIRECTORY_FOR_GO = 'src/github.com/eclipse/codewind-installer'
        DEFAULT_WORKSPACE_DIR_FILE = 'temp_default_dir'
        CODECOV_TOKEN = credentials('codecov-token')
    }

    stages {

        stage ('Build') {
            // This when clause disables Tagged build
            when {
                beforeAgent true
                not {
                    buildingTag()
                }
            }

            steps {
                container('go') {
                    sh '''#!/bin/bash

                        echo "starting preInstall.....: GOPATH=$GOPATH"

                        # add the base directory to the gopath
                        DEFAULT_CODE_DIRECTORY=$PWD
                        cd ../..
                        export GOPATH=$GOPATH:$(pwd)

                        # create a new directory to store the code for go compile
                        if [ -d $CODE_DIRECTORY_FOR_GO ]; then
                            rm -rf $CODE_DIRECTORY_FOR_GO
                        fi
                        mkdir -p $CODE_DIRECTORY_FOR_GO
                        cd $CODE_DIRECTORY_FOR_GO

                        # copy the code into the new directory for go compile
                        cp -r $DEFAULT_CODE_DIRECTORY/* .
                        echo $DEFAULT_CODE_DIRECTORY >> $DEFAULT_WORKSPACE_DIR_FILE

                        # get dep and run it
                        wget -O - https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
                        dep status
                        dep ensure -v

                        # go cache setup
                        mkdir .cache
                        cd .cache
                        mkdir go-build
                        cd ../

                        # now compile the code
                        cd cmd/cli
                        export HOME=$JENKINS_HOME
                        export GOCACHE=/home/jenkins/agent/$CODE_DIRECTORY_FOR_GO/.cache/go-build
                        export GOARCH=amd64
                        GOOS=darwin go build -ldflags="-s -w" -o cwctl-macos
                        GOOS=windows go build -ldflags="-s -w" -o cwctl-win.exe
                        CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cwctl-linux
                        GOOS=linux GOARCH=ppc64le go build -o cwctl-ppc64le
                        chmod -v +x cwctl-*

                        # clean up the cache directory
                        cd ../../
                        rm -rf .cache
                        cd cmd/cli

                        # move the built binaries to the top level direcotory
                        mv cwctl-* ../../
                        cd ../../
                    '''
                }
            }
        }

        stage('Test') {
            // This when clause disables Tagged build
            when {
                beforeAgent true
                not {
                    buildingTag()
                }
            }

            options {
                timeout(time: 10, unit: 'MINUTES')
                retry(3)
            }

            steps {
                echo 'Starting tests'

                container('go') {
                   sh '''#!/bin/bash
                        export GOPATH=/go:/home/jenkins/agent

                        # go cache setup
                        mkdir .cache
                        cd .cache
                        mkdir go-build
                        cd ../
                        export GOCACHE=/home/jenkins/agent/$CODE_DIRECTORY_FOR_GO/.cache/go-build

                        cd ../../$CODE_DIRECTORY_FOR_GO
                        go test ./... -short -coverprofile=coverage.txt -covermode=count
                        TEST_RESULT=$?
                        if [ $TEST_RESULT -ne 0 ]; then
                            exit $TEST_RESULT
                        fi

                        # Report coverage
                        if [ -n "$CODECOV_TOKEN" ]; then
                            echo "Reporting coverage to codecov"
                            bash <(curl -s https://codecov.io/bash) -f ./coverage.txt
                        else
                            echo "CODECOV_TOKEN not set, not reporting coverage"
                        fi

                        # clean up the cache directory
                        rm -rf .cache
                    '''
                }
                echo 'End of test stage'
            }
        }

        stage('Upload') {
            // This when clause disables Tagged build
            when {
                beforeAgent true
                not {
                    buildingTag()
                }
            }

            steps {
                script {
                    sh '''
                        # switch to the code go directory
                        cd ../../$CODE_DIRECTORY_FOR_GO
                        echo $(pwd)
                        if [ -d codewind-installer ]; then
                            rm -rf codewind-installer
                        fi
                        mkdir codewind-installer

                        TIMESTAMP="$(date +%F-%H%M)"
                        # WINDOWS EXE: Submit Windows unsigned.exe and save signed output to signed.exe

                        # only sign windows exe if not a pull request
                        if [ -z $CHANGE_ID ]; then
                            curl -o codewind-installer/cwctl-win-${TIMESTAMP}.exe  -F file=@cwctl-win.exe http://build.eclipse.org:31338/winsign.php
                            rm cwctl-win.exe
                        fi
                        # move other executable to codewind-installer directory and add timestamp to the name
                        for fileid in cwctl-*; do
                            mv -v $fileid codewind-installer/${fileid}-$TIMESTAMP
                        done

                        DEFAULT_WORKSPACE_DIR=$(cat $DEFAULT_WORKSPACE_DIR_FILE)
                        mkdir $DEFAULT_WORKSPACE_DIR/codewind-installer
                        cp -r codewind-installer/* $DEFAULT_WORKSPACE_DIR/codewind-installer
                    '''
                    // stash the executables so they are avaialable outside of this agent
                    dir('codewind-installer') {
                        sh 'echo "Stashing: $(ls -lA cwctl*)"'
                        stash includes: 'cwctl*', name: 'EXECUTABLES'
                    }
                }
            }
        }
        stage('Deploy') {
            // This when clause disables PR/Tag build uploads; you may comment this out if you want your build uploaded.
            when {
                beforeAgent true
                allOf {
                    not {
                        changeRequest()
                    }
                    not {
                        buildingTag()
                    }
                }
            }

            options {
                timeout(time: 120, unit: 'MINUTES')
                retry(3)
            }

            agent any
               steps {
                   sshagent ( ['projects-storage.eclipse.org-bot-ssh']) {

                println("Deploying codewind-installer to download area...")

                // get the stashed executables
                unstash 'EXECUTABLES'

                sh '''
                    export REPO_NAME="codewind-installer"
                    export DOWNLOAD_AREA_URL="https://download.eclipse.org/codewind/$REPO_NAME"
                    export LATEST_DIR="latest"
                    export BUILD_INFO="build_info.properties"
                    export sshHost="genie.codewind@projects-storage.eclipse.org"
                    export deployDir="/home/data/httpd/download.eclipse.org/codewind/$REPO_NAME"
                    export CWCTL_BASENAME="cwctl"
                    export CWCTL_LINUX="${CWCTL_BASENAME}-linux"
                    export CWCTL_PPC64LE="${CWCTL_BASENAME}-ppc64le"
                    export CWCTL_MACOS="${CWCTL_BASENAME}-macos"
                    export CWCTL_WIN="${CWCTL_BASENAME}-win"

                    UPLOAD_DIR="$GIT_BRANCH/$BUILD_ID"
                    BUILD_URL="$DOWNLOAD_AREA_URL/$UPLOAD_DIR"

                    ssh $sshHost rm -rf $deployDir/${UPLOAD_DIR}
                    ssh $sshHost mkdir -p $deployDir/${UPLOAD_DIR}

                    ssh $sshHost rm -rf $deployDir/$GIT_BRANCH/$LATEST_DIR
                    ssh $sshHost mkdir -p $deployDir/$GIT_BRANCH/$LATEST_DIR

                    ls -lA

                    # Copy artifacts before renaming to the build ID'd directory
                    scp ${CWCTL_BASENAME}* $sshHost:$deployDir/${UPLOAD_DIR}

                    # Now prepare to copy to latest/ directory

                    # Rename to remove timestamps
                    mv $CWCTL_LINUX-*    $CWCTL_LINUX
                    mv $CWCTL_PPC64LE-*  $CWCTL_PPC64LE
                    mv $CWCTL_MACOS-*    $CWCTL_MACOS
                    mv $CWCTL_WIN-*      $CWCTL_WIN.exe

                    # Make a targz copy of each build and copy it into zips/
                    export ZIPS_DIR="zips"
                    if [ -d ${ZIPS_DIR} ]; then
                        rm -rf ${ZIPS_DIR}
                    fi

                    mkdir ${ZIPS_DIR}
                    for f in $(ls ${CWCTL_BASENAME}*); do
                        export TARGZ_NAME="${f}.tar.gz"
                        tar -czvf $TARGZ_NAME $f
                        mv $TARGZ_NAME ${ZIPS_DIR}
                    done

                    # Assemble build_info.properties with build date and shasums

                    echo "# Build date: $(date +%F-%T)" >> $BUILD_INFO
                    echo "build_info.url=$BUILD_URL" >> $BUILD_INFO

                    SHA1_LINUX=$(sha1sum $CWCTL_LINUX | cut -d ' ' -f 1)
                    echo "build_info.linux.SHA-1=${SHA1_LINUX}" >> $BUILD_INFO

                    SHA1_PPC64LE=$(sha1sum $CWCTL_PPC64LE | cut -d ' ' -f 1)
                    echo "build_info.ppc64le.SHA-1=${SHA1_PPC64LE}" >> $BUILD_INFO

                    SHA1_MACOS=$(sha1sum $CWCTL_MACOS | cut -d ' ' -f 1)
                    echo "build_info.macos.SHA-1=${SHA1_MACOS}" >> $BUILD_INFO

                    SHA1_WIN=$(sha1sum $CWCTL_WIN.exe | cut -d ' ' -f 1)
                    echo "build_info.win.SHA-1=${SHA1_WIN}" >> $BUILD_INFO

                    # Copy the build.properties, the zips, and the renamed artifacts to the latest/ build directory
                    scp -r ${BUILD_INFO} ${CWCTL_BASENAME}* zips/ $sshHost:$deployDir/$GIT_BRANCH/$LATEST_DIR
                  '''
               }
           }
        }

    }

    post {
        success {
            echo 'Build SUCCESS'
        }
    }
}
