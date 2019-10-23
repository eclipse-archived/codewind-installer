package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreFileOrDirectory(t *testing.T) {
	tests := map[string]struct {
		name            string
		isDir           bool
		shouldBeIgnored bool
	}{
		"success case: directory called node_modules should be ignored": {
			name:            "node_modules",
			isDir:           true,
			shouldBeIgnored: true,
		},
		"success case: directory called load-test-23498729 should be ignored": {
			name:            "load-test-23498729",
			isDir:           true,
			shouldBeIgnored: true,
		},
		"success case: directory called not-a-load-test-23498729 should be ignored": {
			name:            "not-a-load-test-23498729",
			isDir:           true,
			shouldBeIgnored: false,
		},
		"success case: directory called noddy_modules should not be ignored": {
			name:            "noddy_modules",
			isDir:           true,
			shouldBeIgnored: false,
		},
		"success case: file called .DS_Store should be ignored": {
			name:            ".DS_Store",
			isDir:           false,
			shouldBeIgnored: true,
		},
		"success case: file called something.swp should be ignored": {
			name:            "something.swp",
			isDir:           false,
			shouldBeIgnored: true,
		},
		"success case: file called something.swpnot should not be ignored": {
			name:            "something.swpnot",
			isDir:           false,
			shouldBeIgnored: false,
		},
		"success case: file called node_modules should not be ignored": {
			name:            "node_modules",
			isDir:           false,
			shouldBeIgnored: false,
		},
		"success case: directory called .DS_Store should not be ignored": {
			name:            ".DS_Store",
			isDir:           true,
			shouldBeIgnored: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fileIsIgnored := ignoreFileOrDirectory(test.name, test.isDir)

			assert.IsType(t, test.shouldBeIgnored, fileIsIgnored, "Got: %s", fileIsIgnored)

			assert.Equal(t, test.shouldBeIgnored, fileIsIgnored, "fileIsIgnored was %b but should have been %b", fileIsIgnored, test.shouldBeIgnored)
		})
	}
}
