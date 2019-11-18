package config

import (
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/stretchr/testify/assert"
)

func TestPFEOriginError(t *testing.T) {
	tests := map[string]struct {
		connectionID    string
		wantedErrorOp   string
		wantedErrorDesc string
	}{
		"get templates of all styles": {
			connectionID:    "123123123",
			wantedErrorOp:   "config_connection_notfound",
			wantedErrorDesc: "Connection 123123123 not found",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := PFEOrigin(test.connectionID)
			assert.Equal(t, "", got)
			assert.Equal(t, err.Op, test.wantedErrorOp)
			assert.Equal(t, err.Desc, test.wantedErrorDesc)
		})
	}
}

func TestPFEOriginFromConnection(t *testing.T) {
	tests := map[string]struct {
		connection    *connections.Connection
		url           string
		wantedErrorOp string
	}{
		"get templates of all styles": {
			connection: &connections.Connection{
				ID:       "123",
				Label:    "connection1",
				URL:      "ibm.com",
				AuthURL:  "ibm.com/auth",
				Realm:    "codewind",
				ClientID: "codewind",
				Username: "developer",
			},
			url:           "ibm.com",
			wantedErrorOp: "config_connection_notfound",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := PFEOriginFromConnection(test.connection)
			assert.Equal(t, test.url, got)
			assert.Nil(t, err)
		})
	}
}
