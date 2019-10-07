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
    image: golang:1.11-stretch
    tty: true
    command:
    - cat
    resources:
      limits:
        memory: "2Gi"
        cpu: "1"
      requests:
        memory: "2Gi"
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
    }

    stages {

        stage ('Build') {
            steps {
                container('go') {
                    sh '''
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

                        # now compile the code
                        export HOME=$JENKINS_HOME
                        export GOCACHE="off"
                        export GOARCH=amd64
                        GOOS=darwin go build -ldflags="-s -w" -o cwctl-macos
                        GOOS=windows go build -ldflags="-s -w" -o cwctl-win.exe
                        CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cwctl-linux
                        chmod -v +x cwctl-*

                    '''
                }
            }
        }

        stage('Test') {
            agent {
                docker {
                image 'golang:1.11-stretch'
                }
            }
            steps {

                        sh '''#!/usr/bin/env bash
                            echo start pfe containers

                            ARCH=`uname -m`;
                            printf "\n\n${MAGENTA}Platform: $ARCH ${RESET}\n"
                            
                            echo `uname -s`
                            echo `uname -m`

                            # Install docker-compose 
                            curl -L https://github.com/docker/compose/releases/download/1.21.2/docker-compose-`uname -s`-`uname -m` -o ./docker-compose
                            chmod +x ./docker-compose

                            # Create codewind-workspace if it does not exist
                            printf "\n\nCreating codewind-workspace\n"
                            mkdir -m 777 -p codewind-workspace

                            export REPOSITORY='';
                            export TAG
                            export WORKSPACE_DIRECTORY=$PWD/codewind-workspace;
                            export WORKSPACE_VOLUME=cw-workspace;
                            export HOST_OS=$(uname);
                            export REMOTE_MODE;
                            export HOST_HOME=$HOME
                            export ARCH=$(uname -m);
                            # Select the right images for this architecture.
                            if [ "$ARCH" = "x86_64" ]; then
                            export PLATFORM="-amd64"
                            else
                            export PLATFORM="-$ARCH"
                            fi

                            git clone https://github.com/eclipse/codewind.git
                            cd codewind

                            # Start codewind running
                            ../docker-compose -f docker-compose.yaml -f docker-compose-remote.yaml up -d;
                            if [ $? -eq 0 ]; then
                                # Reset so we don't get conflicts
                                unset REPOSITORY;
                                unset WORKSPACE_DIRECTORY;
                                unset REMOTE_MODE;
                                printf "\n\n${GREEN}SUCCESSFULLY STARTED CONTAINERS $RESET\n";
                                printf "\nCurrent running codewind containers\n";
                                docker ps --filter name=codewind
                            else
                                printf "\n\n${RED}FAILED TO START CONTAINERS $RESET\n";
                                exit;
                            fi

                            printf "\n\nPausing for 20 seconds to allow containers to start\n";
                            sleep 20;

                            # Check to see if any containers exited straight away
                            printf "\n\n${BLUE}CHECKING FOR codewind CONTAINERS THAT EXITED STRAIGHT AFTER BEING RUN $RESET\n";
                            EXITED_PROCESSES=$(docker ps -q --filter "name=codewind" --filter "status=exited"  | wc -l)
                            if [ $EXITED_PROCESSES -gt 0 ]; then
                            printf "\n${RED}Exited containers found $RESET\n";
                            # docker ps --filter "name=codewind" --filter "status=exited";
                            NUM_CODE_ZERO=$(docker ps -q --filter "name=codewind" --filter "status=exited" --filter "exited=0" | wc -l);
                            NUM_CODE_ONE=$(docker ps -q --filter "name=codewind" --filter "status=exited" --filter "exited=1" | wc -l);
                            if [ $NUM_CODE_ZERO -gt 0 ]; then
                                printf "\n${RED}$NUM_CODE_ZERO found with an exit code '0' $RESET\n";
                                docker ps --filter "name=codewind" --filter "status=exited" --filter "exited=0";
                                printf "\nUse 'docker logs [container name]' to find why the exit happened";
                            fi
                            if [ $NUM_CODE_ONE -gt 0 ]; then
                                printf "\n${RED}$NUM_CODE_ONE found with an exit code '1' $RESET\n";
                                docker ps --filter "name=codewind" --filter "status=exited" --filter "exited=1";
                                printf "\nUse 'docker logs [container name]' to debug exit";
                            fi
                            else
                            printf "\nNo containers exited \n";
                            fi
                            printf "\n\ncodewind now available\n";
                        
                        '''

            // We need the docker agent for test step so we can run Codewind in docker

                sh '''#!/usr/bin/env bash
                    echo testing script
                    echo $HOME
                    echo $JENKINS_HOME

                    export HOME=$JENKINS_HOME
                    export GOCACHE="off"

                    docker -u $(getent passwd genie.codewind | cut -d: -f3):$(getent passwd genie.codewind | cut -d: -f4)

                    echo "starting preInstall for test stage.....: GOPATH=$GOPATH"

                    # add the base directory to the gopath
                    DEFAULT_CODE_DIRECTORY=$PWD

                    echo file permissions on defaul_code_directory

                    export GOPATH=$GOPATH:$(pwd)

                    # create a new directory to store the code for go compile
                    if [ -d $CODE_DIRECTORY_FOR_GO ]; then
                        rm -rf $CODE_DIRECTORY_FOR_GO
                    fi
                    mkdir -p $CODE_DIRECTORY_FOR_GO
                    cd $CODE_DIRECTORY_FOR_GO

                    # copy the code into the new directory for go compile
                    cp -r $DEFAULT_CODE_DIRECTORY/*.* .
                    cp -r $DEFAULT_CODE_DIRECTORY/Jenkinsfile $DEFAULT_CODE_DIRECTORY/actions $DEFAULT_CODE_DIRECTORY/apiroutes $DEFAULT_CODE_DIRECTORY/config $DEFAULT_CODE_DIRECTORY/errors $DEFAULT_CODE_DIRECTORY/resources $DEFAULT_CODE_DIRECTORY/utils .
                    echo $DEFAULT_CODE_DIRECTORY >> $DEFAULT_WORKSPACE_DIR_FILE

                    
                    # get dep and run it
                    wget -O - https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
                    dep status
                    dep ensure -v

                    # run unit tests
                    echo running unit tests
                    # cd /home/jenkins/$CODE_DIRECTORY_FOR_GO/utils
                    ls
                    # go test
                    go test ./...
                '''

            }
        }

        stage('Upload') {
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
                        cp -r codewind-installer $DEFAULT_WORKSPACE_DIR
                    '''
                    // stash the executables so they are avaialable outside of this agent
                    dir('codewind-installer') {
                        stash includes: '**', name: 'EXECUTABLES'
                    }
                }
            }
        }
        stage('Deploy') {
            // This when clause disables PR build uploads; you may comment this out if you want your build uploaded.
            when {
                beforeAgent true
                not {
                    changeRequest()
                }
            }

            agent any
               steps {
                   sshagent ( ['projects-storage.eclipse.org-bot-ssh']) {
                println("Deploying codewind-installer to download area...")
                sh '''
                    if [ -d codewind-installer ]; then
                        rm -rf codewind-installer
                    fi
                    mkdir codewind-installer
                '''
                // get the stashed executables
                 dir ('codewind-installer') {
                     unstash 'EXECUTABLES'
                 }
                sh '''
                    export REPO_NAME="codewind-installer"
                    export OUTPUT_DIR="$WORKSPACE/dev/ant_build/artifacts"
                    export DOWNLOAD_AREA_URL="https://download.eclipse.org/codewind/$REPO_NAME"
                    export LATEST_DIR="latest"
                    export BUILD_INFO="build_info.properties"
                    export sshHost="genie.codewind@projects-storage.eclipse.org"
                    export deployDir="/home/data/httpd/download.eclipse.org/codewind/$REPO_NAME"
                    export CWCTL_LINUX="cwctl-linux"
                    export CWCTL_MACOS="cwctl-macos"
                    export CWCTL_WIN="cwctl-win"
                    
                    WORKSPACE=$PWD
   
                    ls -la ${WORKSPACE}/$REPO_NAME/*
                    
                    UPLOAD_DIR="$GIT_BRANCH/$BUILD_ID"
                    BUILD_URL="$DOWNLOAD_AREA_URL/$UPLOAD_DIR"
                    
                    ssh $sshHost rm -rf $deployDir/${UPLOAD_DIR}
                    ssh $sshHost mkdir -p $deployDir/${UPLOAD_DIR}

                    ssh $sshHost rm -rf $deployDir/$GIT_BRANCH/$LATEST_DIR
                    ssh $sshHost mkdir -p $deployDir/$GIT_BRANCH/$LATEST_DIR

                    scp ${WORKSPACE}/$REPO_NAME/* $sshHost:$deployDir/${UPLOAD_DIR}

                    mv ${WORKSPACE}/$REPO_NAME/$CWCTL_LINUX-* ${WORKSPACE}/$REPO_NAME/$CWCTL_LINUX
                    mv ${WORKSPACE}/$REPO_NAME/$CWCTL_MACOS-* ${WORKSPACE}/$REPO_NAME/$CWCTL_MACOS
                    mv ${WORKSPACE}/$REPO_NAME/$CWCTL_WIN-* ${WORKSPACE}/$REPO_NAME/$CWCTL_WIN.exe
                    
                    echo "# Build date: $(date +%F-%T)" >> ${WORKSPACE}/$REPO_NAME/$BUILD_INFO
                    echo "build_info.url=$BUILD_URL" >> ${WORKSPACE}/$REPO_NAME/$BUILD_INFO
                    SHA1_LINUX=$(sha1sum ${WORKSPACE}/$REPO_NAME/$CWCTL_LINUX | cut -d ' ' -f 1)
                    echo "build_info.linux.SHA-1=${SHA1_LINUX}" >> ${WORKSPACE}/$REPO_NAME/$BUILD_INFO

                    SHA1_MACOS=$(sha1sum ${WORKSPACE}/$REPO_NAME/$CWCTL_MACOS | cut -d ' ' -f 1)
                    echo "build_info.macos.SHA-1=${SHA1_MACOS}" >> ${WORKSPACE}/$REPO_NAME/$BUILD_INFO

                    SHA1_WIN=$(sha1sum ${WORKSPACE}/$REPO_NAME/$CWCTL_WIN.exe | cut -d ' ' -f 1)
                    echo "build_info.win.SHA-1=${SHA1_WIN}" >> ${WORKSPACE}/$REPO_NAME/$BUILD_INFO

                    scp -r ${WORKSPACE}/$REPO_NAME/* $sshHost:$deployDir/$GIT_BRANCH/$LATEST_DIR
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