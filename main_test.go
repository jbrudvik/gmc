package main_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const executableName string = "gmc"
const helpOutput string = "NAME:\n" +
	"   gmc - (Go mod create) creates Go modules so you can start coding ASAP\n" +
	"\n" +
	"USAGE:\n" +
	"   gmc [global options] [module name]\n" +
	"\n" +
	"VERSION:\n" +
	"   (devel)\n" +
	"\n" +
	"DESCRIPTION:\n" +
	"   `gmc [module name]` creates a directory containing:\n" +
	"   - Go module metadata: go.mod\n" +
	"   - A place to start writing code: main.go\n" +
	"   - A .gitignore file\n" +
	"   \n" +
	"   This module can be immediately run:\n" +
	"   \n" +
	"       $ go run .\n" +
	"       hello, world!\n" +
	"   \n" +
	"   Optionally, the directory can also include:\n" +
	"   - Git repository setup with .gitignore, README.md\n" +
	"   \n" +
	"   More information: https://github.com/jbrudvik/gmc\n" +
	"\n" +
	"GLOBAL OPTIONS:\n" +
	"   --git, -g      create as Git repository (default: false)\n" +
	"   --quiet, -q    silence output (default: false)\n" +
	"   --help, -h     show help (default: false)\n" +
	"   --version, -v  print the version (default: false)\n"

type executableTestCase struct {
	args             []string
	expectedExitCode int
	expectedStdout   string
	expectedStderr   string
}

func TestExecutable(t *testing.T) {
	// Test cases
	tests := []executableTestCase{
		{
			args:             nil,
			expectedExitCode: 1,
			expectedStdout:   helpOutput,
			expectedStderr:   "Error: Module name is required\n\n",
		},
		{
			args:             []string{"-h"},
			expectedExitCode: 0,
			expectedStdout:   helpOutput,
			expectedStderr:   "",
		},
	}
	for _, tc := range tests {
		testName := strings.Join(tc.args, " ")
		t.Run(testName, func(t *testing.T) {
			runExecutableTestCase(t, tc)
		})
	}
}

func runExecutableTestCase(t *testing.T, tc executableTestCase) {
	// Create a temporary test dir (automatically cleaned up)
	tempTestDir := t.TempDir()

	// Build executable
	buildCmd := exec.Command("go", "build", "-o", tempTestDir)
	err := buildCmd.Run()
	if err != nil {
		t.Fatalf("Unable to `go build` %s: $ %s\n", executableName, buildCmd)
	}
	executablePath := filepath.Join(tempTestDir, executableName)

	// Run executable and test outputs
	cmd := exec.Command(executablePath, tc.args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.Equal(t, tc.expectedExitCode, exitError.ExitCode(), "exit code")
		} else {
			t.Fatalf("Unable to parse exit code from: %s", err.Error())
		}
	}
	assert.Equal(t, tc.expectedStdout, stdout.String(), "stdout")
	assert.Equal(t, tc.expectedStderr, stderr.String(), "stderr")
}
