package config

import (
	"testing"

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
