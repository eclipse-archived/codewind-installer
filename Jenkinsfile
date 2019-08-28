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
                        GOOS=darwin go build -ldflags="-s -w" -o codewind-installer-macos
                        GOOS=windows go build -ldflags="-s -w" -o codewind-installer-win.exe
                        GOOS=linux go build -ldflags="-s -w" -o codewind-installer-linux
                        chmod -v +x codewind-installer-*

                    '''
                }
            }
        }

        stage('Test') {
            steps {
                echo 'Testing to be defined.'
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
                            curl -o codewind-installer/codewind-installer-win-${TIMESTAMP}.exe  -F file=@codewind-installer-win.exe http://build.eclipse.org:31338/winsign.php
                            rm codewind-installer-win.exe
                        fi
                        # move other executable to codewind-installer directoryand add timestamp to the name
                        for fileid in codewind-installer-*; do
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
                    export INSTALLER_LINUX="codewind-installer-linux"
                    export INSTALLER_MACOS="codewind-installer-macos"
                    export INSTALLER_WIN="codewind-installer-win"
                    
                    WORKSPACE=$PWD
   
                    ls -la ${WORKSPACE}/$REPO_NAME/*
                    
                    UPLOAD_DIR="$GIT_BRANCH/$BUILD_ID"
                    BUILD_URL="$DOWNLOAD_AREA_URL/$UPLOAD_DIR"
                    
                    ssh $sshHost rm -rf $deployDir/${UPLOAD_DIR}
                    ssh $sshHost mkdir -p $deployDir/${UPLOAD_DIR}

                    ssh $sshHost rm -rf $deployDir/$GIT_BRANCH/$LATEST_DIR
                    ssh $sshHost mkdir -p $deployDir/$GIT_BRANCH/$LATEST_DIR

                    scp ${WORKSPACE}/$REPO_NAME/* $sshHost:$deployDir/${UPLOAD_DIR}

                    mv ${WORKSPACE}/$REPO_NAME/$INSTALLER_LINUX-* ${WORKSPACE}/$REPO_NAME/$INSTALLER_LINUX
                    mv ${WORKSPACE}/$REPO_NAME/$INSTALLER_MACOS-* ${WORKSPACE}/$REPO_NAME/$INSTALLER_MACOS
                    mv ${WORKSPACE}/$REPO_NAME/$INSTALLER_WIN-* ${WORKSPACE}/$REPO_NAME/$INSTALLER_WIN.exe

                    echo "build_info.url=$BUILD_URL" >> ${WORKSPACE}/$REPO_NAME/$BUILD_INFO
                    SHA1_LINUX=$(sha1sum ${WORKSPACE}/$REPO_NAME/$INSTALLER_LINUX | cut -d ' ' -f 1)
                    echo "build_info.linux.SHA-1=${SHA1_LINUX}" >> ${WORKSPACE}/$REPO_NAME/$BUILD_INFO

                    SHA1_MACOS=$(sha1sum ${WORKSPACE}/$REPO_NAME/$INSTALLER_MACOS | cut -d ' ' -f 1)
                    echo "build_info.macos.SHA-1=${SHA1_MACOS}" >> ${WORKSPACE}/$REPO_NAME/$BUILD_INFO

                    SHA1_WIN=$(sha1sum ${WORKSPACE}/$REPO_NAME/$INSTALLER_WIN.exe | cut -d ' ' -f 1)
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
