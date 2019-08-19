package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplates(t *testing.T) {
	tests := map[string]struct {
		wantedType   []Template
		wantedLength int
		wantedErr    error
	}{
		"success case: no input": {
			wantedType:   []Template{},
			wantedLength: 8,
			wantedErr:    nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplates()
			assert.IsType(t, test.wantedType, got)
			assert.Equal(t, test.wantedLength, len(got))
			assert.IsType(t, test.wantedErr, err)
		})
	}
}
