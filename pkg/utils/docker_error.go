package utils

import "encoding/json"

type DockerError struct {
	Op   string
	Err  error
	Desc string
}

const (
	errOpValidate = "docker_validate" // validate docker images
)

const (
	textBadDigest = "Failed to validate docker image checksum"
)

// DockerError : Error formatted in JSON containing an errorOp and a description
func (de *DockerError) Error() string {
	type Output struct {
		Operation   string `json:"error"`
		Description string `json:"error_description"`
	}
	tempOutput := &Output{Operation: de.Op, Description: de.Err.Error()}
	jsonError, _ := json.Marshal(tempOutput)
	return string(jsonError)
}
