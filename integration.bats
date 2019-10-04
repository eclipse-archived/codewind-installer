#!/usr/bin/env bats

@test "invoke install command - install latest with --json" {
  run go run main.go install --json
  echo "status = ${status}"
  echo "output trace = ${output}"
    [ "$status" -eq 0 ]
}

@test "invoke status -j command - output = '{"status":"stopped","installed-versions":["latest"]}'" {
  run go run main.go status -j
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"status":"stopped","installed-versions":["latest"]}' ]
  [ "$status" -eq 0 ]
}

@test "invoke start command - Start dockerhub images (latest)" {
  run go run main.go start -t latest
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$status" -eq 0 ]
}

@test "invoke stop-all command - Stop dockerhub images (latest)" {
  run go run main.go stop-all
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$status" -eq 0 ]
}

@test "invoke remove command - remove all dockerhub images" {
  run go run main.go remove
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$status" -eq 0 ]
}

############################
# Deployment command tests #
############################

@test "invoke dep reset command - reset deployments file" {
  run go run main.go dep reset
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"status":"OK","status_message":"Deployment list reset"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke dep list command - contains just 1 local deployment" {
  run go run main.go dep list
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"schemaversion":1,"active":"local","deployments":[{"id":"local","label":"Codewind local deployment","url":"","auth":"","realm":"","clientid":""}]}' ]
   [ "$status" -eq 0 ]
}

@test "invoke dep add command - add new deployment to the list" {
  run go run main.go dep add -id kube --label "kube-cluster" --url http://mykube:12345 --auth http://myauth:12345 --realm codewind-cloud --clientid codewind
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"status":"OK","status_message":"Deployment added"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke dep list command - ensure both deployments exist " {
  run go run main.go dep list
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"schemaversion":1,"active":"local","deployments":[{"id":"local","label":"Codewind local deployment","url":"","auth":"","realm":"","clientid":""},{"id":"kube","label":"kube-cluster","url":"http://mykube:12345","auth":"http://myauth:12345","realm":"codewind-cloud","clientid":"codewind"}]}' ]
   [ "$status" -eq 0 ]
}

@test "invoke dep target command - set a target to something unknown" {
  run go run main.go dep target -id noname
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"error":"dep_not_found","error_description":"Target deployment not found"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke dep target command - set the target to kube" {
  run go run main.go dep target --id kube
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"status":"OK","status_message":"New target set"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke dep target command - check the target is now kube" {
  run go run main.go dep target
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"id":"kube","label":"kube-cluster","url":"http://mykube:12345","auth":"http://myauth:12345","realm":"codewind-cloud","clientid":"codewind"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke dep remove command - delete target kube" {
  run go run main.go dep remove --id kube
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"status":"OK","status_message":"Deployment removed"}' ]
   [ "$status" -eq 0 ]
}

@test "invoke dep target command - check target returns to local" {
  run go run main.go dep target
  echo "status = ${status}"
  echo "output trace = ${output}"
   [ "$output" = '{"id":"local","label":"Codewind local deployment","url":"","auth":"","realm":"","clientid":""}' ]
   [ "$status" -eq 0 ]
}

#########################
# Keyring command tests #
#########################

@test "invoke seckeyring update command - create a key" {
  run go run main.go seckeyring update --depid local --username testuser --password secretphrase
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"status":"OK"}' ]
  [ "$status" -eq 0 ]
}

@test "invoke seckeyring update command - update a key" {
  run go run main.go seckeyring update --depid local --username testuser --password new_secretphrase
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"status":"OK"}' ]
  [ "$status" -eq 0 ]
}

@test "invoke seckeyring validate command - validate a key" {
  run go run main.go seckeyring validate --depid local --username testuser
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"status":"OK"}' ]
  [ "$status" -eq 0 ]
}

@test "invoke seckeyring validate command - key not found (incorrect deployment)" {
  run go run main.go seckeyring validate --depid remoteNotKnown --username testuser
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"error":"sec_keyring","error_description":"secret not found in keyring"}' ]
  [ "$status" -eq 0 ]
}

@test "invoke seckeyring validate command - key not found (incorrect username)"  {
  run go run main.go seckeyring validate --depid local --username testuser_unknown
  echo "status = ${status}"
  echo "output trace = ${output}"
  [ "$output" = '{"error":"sec_keyring","error_description":"secret not found in keyring"}' ]
  [ "$status" -eq 0 ]
}

