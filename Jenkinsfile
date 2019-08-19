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
					sh 'echo "starting preInstall.....: GOPATH=$GOPATH"'
					sh '''
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
					WORKSPACE=$PWD
					ls -la ${WORKSPACE}/codewind-installer/*
					if [ -z $CHANGE_ID ]; then
    					UPLOAD_DIR="$GIT_BRANCH/$BUILD_ID"
					else
    					UPLOAD_DIR="pr/$CHANGE_ID/$BUILD_ID"
					fi
                 	ssh genie.codewind@projects-storage.eclipse.org rm -rf /home/data/httpd/download.eclipse.org/codewind/codewind-installer/${UPLOAD_DIR}
            		ssh genie.codewind@projects-storage.eclipse.org mkdir -p /home/data/httpd/download.eclipse.org/codewind/codewind-installer/${UPLOAD_DIR}
                    scp -r ${WORKSPACE}/codewind-installer/* genie.codewind@projects-storage.eclipse.org:/home/data/httpd/download.eclipse.org/codewind/codewind-installer/${UPLOAD_DIR}
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
