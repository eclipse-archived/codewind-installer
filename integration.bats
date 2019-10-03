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