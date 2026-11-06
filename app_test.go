package main

import (
	"testing"
)

func TestInferItemMetadata(t *testing.T) {
	app := &App{}

	testCases := []struct {
		input       string
		expectedName string
		expectedDir  string
	}{
		{
			input:       "/home/tpsapp/.local/bin/dicemachine",
			expectedName: "dicemachine",
			expectedDir:  "/home/tpsapp/.local/bin",
		},
		{
			input:       "/usr/bin/google-chrome-stable",
			expectedName: "google-chrome-stable",
			expectedDir:  "/usr/bin",
		},
		{
			input:       "/opt/scripts/deploy.sh",
			expectedName: "deploy",
			expectedDir:  "/opt/scripts",
		},
	}

	for _, tc := range testCases {
		meta := app.InferItemMetadata(tc.input)
		if meta.Name != tc.expectedName {
			t.Errorf("For input %s, expected Name '%s', got '%s'", tc.input, tc.expectedName, meta.Name)
		}
		if meta.WorkingDirectory != tc.expectedDir {
			t.Errorf("For input %s, expected WorkingDirectory '%s', got '%s'", tc.input, tc.expectedDir, meta.WorkingDirectory)
		}
	}
}
