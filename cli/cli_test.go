package cli_test

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/jbrudvik/gmc/cli"
)

// TODO: Should I change all of these constants to []byte?
const helpOutput string = "NAME:\n" +
	"   gmc - (Go mod create) creates Go modules\n" +
	"\n" +
	"USAGE:\n" +
	"   gmc [global options] [module name]\n" +
	"\n" +
	"VERSION:\n" +
	"   v0.0.1\n" +
	"\n" +
	"DESCRIPTION:\n" +
	"   gmc [module name] creates a directory containing:\n" +
	"   - go.mod            Go module metadata\n" +
	"   - .gitignore        ignores your module's binary\n" +
	"   - main.go           your module's first code\n" +
	"   - .nova (Optional)  Nova editor configuration\n" +
	"   \n" +
	"   This directory can be immediately built/run/installed using the `go` CLI.\n" +
	"   \n" +
	"   More information: https://github.com/jbrudvik/gmc\n" +
	"\n" +
	"GLOBAL OPTIONS:\n" +
	"   --nova, -n     include Nova configuration (default: false)\n" +
	"   --help, -h     show help (default: false)\n" +
	"   --version, -v  print the version (default: false)\n"

var versionOutput string = fmt.Sprintf("gmc version %s\n", cli.Version)

const mainGoContents string = "package main\n" +
	"\n" +
	"import (\n" +
	"	\"fmt\"\n" +
	")\n" +
	"\n" +
	"func main() {\n" +
	"	fmt.Println(\"hello!\")\n" +
	"}\n"

const novaTaskContents string = `{
  "actions": {
    "build": {
      "enabled": true,
      "script": "go build && go test ./..."
    },
    "clean": {
      "enabled": true,
      "script": "go clean"
    },
    "run": {
      "enabled": true,
      "script": "go run ."
    }
  },
  "openLogOnRun": "start"
}
`

const errorMessageUnknownFlag string = "Error: Unknown flag\n\n"
const errorMessageModuleNameRequired string = "Error: Module name is required\n\n"
const errorMessageTooManyModuleNames string = "Error: Only one module name is allowed\n\n"

type file struct {
	path string

	// Non-nil -> file, nil -> directory
	content []byte
}

// TODO: Figure out how to ensure there aren't extra files -- maybe just count the tree somehow?
// TODO: Should we look for any other characteristics?

func TestRun(t *testing.T) {
	tests := []struct {
		args                []string
		expectedOutput      string
		expectedErrorOutput string
		expectedExitCode    int
		expectedFiles       []file
	}{
		{
			[]string{"-h"},
			helpOutput,
			"",
			0,
			nil,
		},
		{
			[]string{"--help"},
			helpOutput,
			"",
			0,
			nil,
		},
		{
			[]string{"-v"},
			versionOutput,
			"",
			0,
			nil,
		},
		{
			[]string{"--version"},
			versionOutput,
			"",
			0,
			nil,
		},
		{
			[]string{},
			helpOutput,
			errorMessageModuleNameRequired,
			1,
			nil,
		},
		{
			[]string{"-e"},
			helpOutput,
			errorMessageUnknownFlag,
			1,
			nil,
		},
		{
			[]string{"-e", "a1"},
			helpOutput,
			errorMessageUnknownFlag,
			1,
			nil,
		},
		{
			[]string{"-n"},
			helpOutput,
			errorMessageModuleNameRequired,
			1,
			nil,
		},
		{
			[]string{"--nova"},
			helpOutput,
			errorMessageModuleNameRequired,
			1,
			nil,
		},
		{
			[]string{"a1", "a2"},
			helpOutput,
			errorMessageTooManyModuleNames,
			1,
			nil,
		},
		{
			[]string{"a1"},
			"",
			"",
			0,
			[]file{
				{"a1", nil},
				{"a1/go.mod", []byte("module a1\n\ngo 1.18\n")},
				{"a1/.gitignore", []byte("a1")},
				{"a1/main.go", []byte(mainGoContents)},
			},
		},
		{
			[]string{"-n", "a2"},
			"",
			"",
			0,
			[]file{
				{"a2", nil},
				{"a2/go.mod", []byte("module a2\n\ngo 1.18\n")},
				{"a2/.gitignore", []byte("a2")},
				{"a2/main.go", []byte(mainGoContents)},
				{"a2/.nova", nil},
				{"a2/.nova/Tasks", nil},
				{"a2/.nova/Tasks/Go.json", []byte(novaTaskContents)},
			},
		},
		{
			[]string{"--nova", "a3"},
			"",
			"",
			0,
			[]file{
				{"a3", nil},
				{"a3/go.mod", []byte("module a3\n\ngo 1.18\n")},
				{"a3/.gitignore", []byte("a3")},
				{"a3/main.go", []byte(mainGoContents)},
				{"a3/.nova", nil},
				{"a3/.nova/Tasks", nil},
				{"a3/.nova/Tasks/Go.json", []byte(novaTaskContents)},
			},
		},
		{
			[]string{"example.com/foo"},
			"",
			"",
			0,
			[]file{
				{"foo", nil},
				{"foo/go.mod", []byte("module example.com/foo\n\ngo 1.18\n")},
				{"foo/.gitignore", []byte("foo")},
				{"foo/main.go", []byte(mainGoContents)},
			},
		},
		{
			[]string{"--nova", "example.com/foo/bar"},
			"",
			"",
			0,
			[]file{
				{"bar", nil},
				{"bar/go.mod", []byte("module example.com/foo/bar\n\ngo 1.18\n")},
				{"bar/.gitignore", []byte("bar")},
				{"bar/main.go", []byte(mainGoContents)},
				{"bar/.nova", nil},
				{"bar/.nova/Tasks", nil},
				{"bar/.nova/Tasks/Go.json", []byte(novaTaskContents)},
			},
		},
		{
			[]string{"-n", "example.com/foo/bar/baz"},
			"",
			"",
			0,
			[]file{
				{"baz", nil},
				{"baz/go.mod", []byte("module example.com/foo/bar/baz\n\ngo 1.18\n")},
				{"baz/.gitignore", []byte("baz")},
				{"baz/main.go", []byte(mainGoContents)},
				{"baz/.nova", nil},
				{"baz/.nova/Tasks", nil},
				{"baz/.nova/Tasks/Go.json", []byte(novaTaskContents)},
			},
		},
	}

	// setUpTestDir(t)
	testDir := setUpTestDir(t)
	defer tearDownTestDir(t, testDir)

	for _, tc := range tests {
		input := fmt.Sprintf("%s", tc.args)

		var outputBuffer bytes.Buffer
		var errorOutputBuffer bytes.Buffer
		exitCodeHandler := func(exitCode int) {
			if tc.expectedExitCode != exitCode {
				t.Errorf(testCaseUnexpectedMessage(input, "exit code", strconv.Itoa(tc.expectedExitCode), strconv.Itoa(exitCode)))
			}
		}

		app := cli.AppWithCustomOutputAndExit(&outputBuffer, &errorOutputBuffer, exitCodeHandler)
		_ = app.Run(append([]string{"gmc"}, tc.args...))

		actualOutput := outputBuffer.String()
		actualErrorOutput := errorOutputBuffer.String()
		if actualOutput != tc.expectedOutput {
			t.Error(testCaseUnexpectedMessage(input, "output", tc.expectedOutput, actualOutput))
		}
		if actualErrorOutput != tc.expectedErrorOutput {
			t.Error(testCaseUnexpectedMessage(input, "error output", tc.expectedErrorOutput, actualErrorOutput))
		}

		cwd, err := os.Getwd()
		if err != nil {
			t.Error("Could not get cwd", err)
		}
		for _, f := range tc.expectedFiles {
			absolutePath := path.Join(cwd, f.path)

			// TODO: Ensure there aren't extra files in actual

			// TODO: Do a check to see if the file exists at all?

			if f.content != nil {
				bytes, err := os.ReadFile(absolutePath)
				if err != nil {
					errorMessage := fmt.Sprintf("Test with input %s: Unable to read expected file: %s:\n", input, absolutePath)
					t.Error(errorMessage)
				} else {
					expectedFileContent := string(f.content)
					actualFileContent := string(bytes)
					if expectedFileContent != actualFileContent {
						errorMessage := testCaseUnexpectedMessage(input, fmt.Sprintf("file content at path: %s", f.path), expectedFileContent, actualFileContent)
						t.Error(errorMessage)
					}
				}
			} else {
				// TODO: Compare dir
				entries, err := os.ReadDir(absolutePath)
				if err != nil {
					errorMessage := fmt.Sprintf("Test with input %s: Unable to read expected file: %s:\n", input, absolutePath)
					t.Error(errorMessage)
				} else {
					// TODO: Actually look at the properties / contents of the directory
					fmt.Println(entries)
					// for _, entry := range entries {
					// 	entry.Name()
					// }
					// t.Error()
				}
			}
		}
	}
}

func setUpTestDir(t *testing.T) string {
	testDir, err := os.MkdirTemp(".", "cli_test_dir")
	if err != nil {
		t.Error("Failure during test setup:", err)
	}
	err = os.Chdir(testDir)
	if err != nil {
		t.Error("Failure during test setup:", err)
	}
	return testDir
}

func tearDownTestDir(t *testing.T, testDir string) {
	err := os.Chdir("..")
	if err != nil {
		t.Error("Failure during test teardown:", err)
	}
	err = os.RemoveAll(testDir)
	if err != nil {
		t.Error("Failure during test teardown:", err)
	}
}

// TODO: Rename this method -- maybe do really factor out for my assert library
// TODO: Rename `thing`
func testCaseUnexpectedMessage(input string, thing string, expected string, actual string) string {
	return fmt.Sprintf("Test with input %s: Unexpected %s\nExpected: %s\nActual  : %s\n", input, thing, expected, actual)
}
