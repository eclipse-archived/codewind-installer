#!/usr/bin/env bats

@test "invoke install command - install latest with --json" {
  cd cmd/cli/
  run go run main.go install --json
  echo "status = ${status}"
  echo "output trace = ${output}"
    [ "$status" -eq 0 ]
}

@test "invoke status -j command - output = '{"status":"stopped","installed-versions":["latest"]}'" {
  cd cmd/cli/
  run go run main.go status -j
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"status":"stopped","installed-versions":["latest"]}' ]
  [ "$status" -eq 0 ]
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> fmt.println() used for error objects to stop bad string parsing with logrus
}

@test "invoke start command - Start dockerhub images (latest)" {
  cd cmd/cli/
  run go run main.go start -t latest
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$status" -eq 0 ]
}

@test "invoke stop-all command - Stop dockerhub images (latest)" {
  cd cmd/cli/
  run go run main.go stop-all
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$status" -eq 0 ]
}

@test "invoke remove command - remove all dockerhub images" {
  cd cmd/cli/
  run go run main.go remove
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$status" -eq 0 ]
}

############################
# Connection command tests #
############################

@test "invoke con reset command - reset connections file" {
  cd cmd/cli/
<<<<<<< HEAD
  run go run main.go --json con reset
=======
  run go run main.go con reset
>>>>>>> fmt.println() used for error objects to stop bad string parsing with logrus
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"status":"OK","status_message":"Connection list reset"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke con list command - contains just 1 local connection" {
  cd cmd/cli/
<<<<<<< HEAD
  run go run main.go --json con list
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"schemaversion":1,"connections":[{"id":"local","label":"Codewind local connection","url":"","auth":"","realm":"","clientid":"","username":""}]}' ]
=======
  run go run main.go con list
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"schemaversion":1,"connections":[{"id":"local","label":"Codewind local connection","url":"","auth":"","realm":"","clientid":""}]}' ]
>>>>>>> fmt.println() used for error objects to stop bad string parsing with logrus
   [ "$status" -eq 0 ]
}

@test "invoke con add command - add new connection to the list" {
  skip "environment not available yet"
  cd cmd/cli/
  run go run main.go con add -d kube --label "kube-cluster" --url http://mykube:12345 --auth http://myauth:12345 --realm codewind-cloud --clientid codewind
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"status":"OK","status_message":"Connection added"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke con list command - ensure both connections exist " {
  skip "environment not available yet"
  cd cmd/cli/
  run go run main.go con list
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"schemaversion":1,"connections":[{"id":"local","label":"Codewind local connection","url":"","auth":"","realm":"","clientid":""},{"id":"kube","label":"kube-cluster","url":"http://mykube:12345","auth":"http://myauth:12345","realm":"codewind-cloud","clientid":"codewind"}]}' ]
   [ "$status" -eq 0 ]
}

@test "invoke con target command - set a target to something unknown" {
  skip "environment not available yet"
  cd cmd/cli/
  run go run main.go con target -d noname
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"error":"con_not_found","error_description":"Target connection not found"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke con target command - set the target to kube" {
  skip "environment not available yet"
  cd cmd/cli/
  run go run main.go con target -d kube
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"status":"OK","status_message":"New target set"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke con target command - check the target is now kube" {
  skip "environment not available yet"
  cd cmd/cli/
  run go run main.go con target
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"id":"kube","label":"kube-cluster","url":"http://mykube:12345","auth":"http://myauth:12345","realm":"codewind-cloud","clientid":"codewind"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke con remove command - delete target kube" {
  skip "environment not available yet"
  cd cmd/cli/
  run go run main.go con remove --id kube
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"status":"OK","status_message":"Connection removed"}' ]
   [ "$status" -eq 0 ]
}

#########################
# Keyring command tests #
#########################

@test "invoke seckeyring update command - create a key" {
  cd cmd/cli/
  run go run main.go seckeyring update --conid local --username testuser --password seCretphrase
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"status":"OK"}' ]
  [ "$status" -eq 0 ]
}

@test "invoke seckeyring update command - update a key" {
  cd cmd/cli/
  run go run main.go seckeyring update --conid local --username testuser --password new_secretPhrase
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"status":"OK"}' ]
  [ "$status" -eq 0 ]
}

@test "invoke seckeyring validate command - validate a key" {
  cd cmd/cli/
  run go run main.go seckeyring validate --conid local --username testuser
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"status":"OK"}' ]
  [ "$status" -eq 0 ]
}

@test "invoke seckeyring validate command - key not found (incorrect connection)" {
  cd cmd/cli/
  run go run main.go seckeyring validate --conid remoteNotKnown --username testuser
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "${lines[0]}" = '{"error":"sec_keyring","error_description":"secret not found in keyring"}' ]
  [ "${lines[1]}" = "exit status 1" ]
  [ "$status" -eq 1 ]
}

@test "invoke seckeyring validate command - key not found (incorrect username)"  {
  cd cmd/cli/
  run go run main.go seckeyring validate --conid local --username testuser_unknown
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "${lines[0]}" = '{"error":"sec_keyring","error_description":"secret not found in keyring"}' ]
  [ "${lines[1]}" = "exit status 1" ]
  [ "$status" -eq 1 ]
<<<<<<< HEAD
}

=======
}
>>>>>>> all json output using fmt.println, all other output using logrus
=======
}
>>>>>>> fmt.println() used for error objects to stop bad string parsing with logrus
