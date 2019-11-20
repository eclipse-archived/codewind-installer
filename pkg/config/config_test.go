package config

import (
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/stretchr/testify/assert"
)

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
