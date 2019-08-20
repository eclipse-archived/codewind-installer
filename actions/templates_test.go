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
		"success case": {
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
			assert.Equal(t, test.wantedErr, err)
		})
	}
}

func TestGetTemplateStyles(t *testing.T) {
	tests := map[string]struct {
		want      []string
		wantedErr error
	}{
		"success case": {
			want:      []string{"Codewind"},
			wantedErr: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplateStyles()
			assert.Equal(t, test.want, got)
			assert.IsType(t, test.wantedErr, err)
		})
	}
}
