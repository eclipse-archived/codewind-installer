#!groovy​

pipeline {

	agent any

    options {
        skipStagesAfterUnstable()
    }

	environment {
  		PRODUCT_NAME = 'codewind-installer'
	}

	stages {
		stage ('preBuild') {
			agent { 
				node { 
					label '' 
					customWorkspace 'src/codewind-installer' 
				}
			}

			steps {
				script {
					sh 'echo "starting preInstall.....: GOPATH=$GOPATH"'
					sh '''
						# add the base directory to the gopath
						CODE_DIRECTORY=$PWD
						projectDir=$(basename $PWD)
						cd ../..
						export GOPATH=$GOPATH:$(pwd)
						cd $CODE_DIRECTORY
						# get all of of the go dependences
						//curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
						//dep ensure -v
						echo "Building in directory $(pwd)"
						'''
				}
			}
		}

		stage('Build') {

			agent { 
				node { 
					label '' 
					customWorkspace 'src/codewind-installer' 
				}
			}

			steps {
				script {
					sh '''
						# add the base directory to the gopath
						CODE_DIRECTORY=$PWD
						projectDir=$(basename $PWD)
						cd ../..
						export GOPATH=$GOPATH:$(pwd)
						cd $CODE_DIRECTORY
						//GOOS=darwin go build -o ${PRODUCT_NAME}-macos
  						//GOOS=windows go build -o ${PRODUCT_NAME}-win.exe
  						//GOOS=linux go build -o ${PRODUCT_NAME}-linux
  						//chmod -v +x ${PRODUCT_NAME}-*
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
 			agent { 
				node { 
					label '' 
					customWorkspace 'src/codewind-installer' 
				}
			}

           steps {
				script {
					sh '''
						if [ -d $PRODUCT_NAME ]; then
							rm -rf $PRODUCT_NAME
						fi	
						mkdir $PRODUCT_NAME
						
						# WINDOWS EXE: Submit Windows unsigned.exe and save signed output to signed.exe
                        curl -o $PRODUCT_NAME/${PRODUCT_NAME}-win-signed.exe  -F file=@${PRODUCT_NAME}-win.exe http://build.eclipse.org:31338/winsign.php

  						mv -v $PRODUCT_NAME-* $PRODUCT_NAME/
						echo "zip up the images - does not work!"  
					'''
				}		 
				script { 
					zip archive: true,  dir: 'codewind-installer', glob: ' ', zipFile: 'codewind-installer.zip'
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