package storage

import (
	"bytes"
	"strings"
	"testing"

	"github.com/plutov/formulosity/api/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSaveFile(t *testing.T) {
	tempDir := t.TempDir()
	p := &File{
		uploadDir: tempDir,
	}

	cases := []struct {
		name        string
		file        *types.File
		expectError bool
		expectFile  bool
	}{
		{
			name: "should save file successfully",
			file: &types.File{
				Name: "valid_filename.txt",
				Data: bytes.NewReader([]byte("test data")),
			},
			expectError: false,
			expectFile:  true,
		},
		{
			name: "should return error if file name is invalid",
			file: &types.File{
				Name: "/invalid?filename<>|",
				Data: bytes.NewReader([]byte("test data")),
			},
			expectError: true,
			expectFile:  false,
		},
		{
			name: "should return error if file name is empty",
			file: &types.File{
				Name: "",
				Data: bytes.NewReader([]byte("test data")),
			},
			expectError: true,
			expectFile:  false,
		},
		{
			name: "should return error if file name is too long",
			file: &types.File{
				Name: strings.Repeat("a", 255),
				Data: bytes.NewReader([]byte("test data")),
			},
			expectError: true,
			expectFile:  false,
		},
		{
			name: "should return error if file is empty",
			file: &types.File{
				Name: "valid_filename.txt",
				Data: bytes.NewReader([]byte("")),
			},
			expectError: true,
			expectFile:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := p.SaveFile(tc.file)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, result, tc.file.Name)
				assert.FileExists(t, result)
			}

			if tc.expectFile {
				assert.FileExists(t, result)
			} else {
				assert.NoFileExists(t, result)
			}
		})
	}
}