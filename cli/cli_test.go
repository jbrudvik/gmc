package cli_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/jbrudvik/gmc/cli"
)

type file struct {
	path    string
	content []byte // Non-nil -> file, nil -> directory
	perm    fs.FileMode
}

const dirPerms fs.FileMode = 0755 | fs.ModeDir
const filePerms fs.FileMode = 0644

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
				{"a1", nil, dirPerms},
				{"a1/go.mod", []byte("module a1\n\ngo 1.18\n"), filePerms},
				{"a1/.gitignore", []byte("a1"), filePerms},
				{"a1/main.go", []byte(mainGoContents), filePerms},
			},
		},
		{
			[]string{"-n", "a2"},
			"",
			"",
			0,
			[]file{
				{"a2", nil, dirPerms},
				{"a2/go.mod", []byte("module a2\n\ngo 1.18\n"), filePerms},
				{"a2/.gitignore", []byte("a2"), filePerms},
				{"a2/main.go", []byte(mainGoContents), filePerms},
				{"a2/.nova", nil, dirPerms},
				{"a2/.nova/Tasks", nil, dirPerms},
				{"a2/.nova/Tasks/Go.json", []byte(novaTaskContents), filePerms},
			},
		},
		{
			[]string{"--nova", "a3"},
			"",
			"",
			0,
			[]file{
				{"a3", nil, dirPerms},
				{"a3/go.mod", []byte("module a3\n\ngo 1.18\n"), filePerms},
				{"a3/.gitignore", []byte("a3"), filePerms},
				{"a3/main.go", []byte(mainGoContents), filePerms},
				{"a3/.nova", nil, dirPerms},
				{"a3/.nova/Tasks", nil, dirPerms},
				{"a3/.nova/Tasks/Go.json", []byte(novaTaskContents), filePerms},
			},
		},
		{
			[]string{"example.com/foo"},
			"",
			"",
			0,
			[]file{
				{"foo", nil, dirPerms},
				{"foo/go.mod", []byte("module example.com/foo\n\ngo 1.18\n"), filePerms},
				{"foo/.gitignore", []byte("foo"), filePerms},
				{"foo/main.go", []byte(mainGoContents), filePerms},
			},
		},
		{
			[]string{"--nova", "example.com/foo/bar"},
			"",
			"",
			0,
			[]file{
				{"bar", nil, dirPerms},
				{"bar/go.mod", []byte("module example.com/foo/bar\n\ngo 1.18\n"), filePerms},
				{"bar/.gitignore", []byte("bar"), filePerms},
				{"bar/main.go", []byte(mainGoContents), filePerms},
				{"bar/.nova", nil, dirPerms},
				{"bar/.nova/Tasks", nil, dirPerms},
				{"bar/.nova/Tasks/Go.json", []byte(novaTaskContents), filePerms},
			},
		},
		{
			[]string{"-n", "example.com/foo/bar/baz"},
			"",
			"",
			0,
			[]file{
				{"baz", nil, dirPerms},
				{"baz/go.mod", []byte("module example.com/foo/bar/baz\n\ngo 1.18\n"), filePerms},
				{"baz/.gitignore", []byte("baz"), filePerms},
				{"baz/main.go", []byte(mainGoContents), filePerms},
				{"baz/.nova", nil, dirPerms},
				{"baz/.nova/Tasks", nil, dirPerms},
				{"baz/.nova/Tasks/Go.json", []byte(novaTaskContents), filePerms},
			},
		},
	}

	testDir := setUpTestDir(t)
	defer tearDownTestDir(t, testDir)

	for _, tc := range tests {
		input := fmt.Sprintf("%s", tc.args)

		var outputBuffer bytes.Buffer
		var errorOutputBuffer bytes.Buffer
		exitCodeHandler := func(exitCode int) {
			if tc.expectedExitCode != exitCode {
				t.Errorf(testCaseUnexpectedMessage(input, "exit code", tc.expectedExitCode, exitCode))
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
			// TODO: Do we need to use first value in return? Ideally, we should...
			absolutePath := path.Join(cwd, f.path)
			_, err := expectedFileIsAtPath(f, absolutePath)
			if err != nil {
				errorMessage := testInputUnexpectedMessage(input, err.Error())
				t.Error(errorMessage)
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

func expectedFileIsAtPath(f file, filePath string) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		errorMessage := fmt.Sprintf("Unable to stat expected file: %s", filePath)
		return false, errors.New(errorMessage)
	}

	actualMode := fileInfo.Mode()
	if f.perm != actualMode {
		errorMessage := testResultUnexpectedMessage(fmt.Sprintf("file perms at path: %s", filePath), f.perm, actualMode)
		return false, errors.New(errorMessage)
	}

	if f.content != nil {
		// Compare files
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			errorMessage := fmt.Sprintf("Unable to read expected file: %s:\n", filePath)
			return false, errors.New(errorMessage)
		} else {
			expectedFileContent := string(f.content)
			actualFileContent := string(bytes)
			if expectedFileContent != actualFileContent {
				errorMessage := testResultUnexpectedMessage(fmt.Sprintf("file content at path: %s", filePath), expectedFileContent, actualFileContent)
				return false, errors.New(errorMessage)
			}
		}
	} else {
		// Compare dirs
		entries, err := os.ReadDir(filePath)
		if err != nil {
			errorMessage := fmt.Sprintf("Unable to read expected file: %s:\n", filePath)
			return false, errors.New(errorMessage)
		} else {
			// TODO: Actually look at the properties / contents of the directory -- maybe this is how to see if all files are present (probably!)
			// Will this handle everything?
			fmt.Println(entries)
			// for _, entry := range entries {
			// 	entry.Name()
			// }
			// t.Error()
		}
	}
	return true, nil
}

// TODO: Rename this method
// TODO: Rename `thing`
func testCaseUnexpectedMessage[T any](input string, thing string, expected T, actual T) string {
	testResultMessage := testResultUnexpectedMessage(thing, expected, actual)
	testCaseMessage := testInputUnexpectedMessage(input, testResultMessage)
	return testCaseMessage
}

func testInputUnexpectedMessage(input string, message string) string {
	return fmt.Sprintf("Test with input %s: %s", input, message)
}

func testResultUnexpectedMessage[T any](thing string, expected T, actual T) string {
	return fmt.Sprintf("Unexpected %s\nExpected: %v\nActual  : %v\n", thing, expected, actual)
}
