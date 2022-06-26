package cli_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/jbrudvik/gmc/cli"
)

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

var versionOutput string = fmt.Sprintf("%s version %s\n", cli.Name, cli.Version)

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
	name    string
	perm    fs.FileMode
	content []byte // Non-nil for file
	files   []file // Non-nil for directory
}

const dirPerms fs.FileMode = 0755 | fs.ModeDir
const filePerms fs.FileMode = 0644

func TestRun(t *testing.T) {
	tests := []struct {
		args                []string
		expectedOutput      string
		expectedErrorOutput string
		expectedExitCode    int
		expectedFiles       *file
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
			"Creating Go module \"a1\"...\nSuccess! Created Go module \"a1\" in new directory: a1\n",
			"",
			0,
			&file{"a1", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module a1\n\ngo 1.18\n"), nil},
				{".gitignore", filePerms, []byte("a1"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
			}},
		},
		{
			[]string{"-n", "a2"},
			"Creating Go module \"a2\"...\nSuccess! Created Go module \"a2\" in new directory: a2\n",
			"",
			0,
			&file{"a2", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module a2\n\ngo 1.18\n"), nil},
				{".gitignore", filePerms, []byte("a2"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
		},
		{
			[]string{"--nova", "a3"},
			"Creating Go module \"a3\"...\nSuccess! Created Go module \"a3\" in new directory: a3\n",
			"",
			0,
			&file{"a3", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module a3\n\ngo 1.18\n"), nil},
				{".gitignore", filePerms, []byte("a3"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
		},
		{
			[]string{"example.com/foo"},
			"Creating Go module \"example.com/foo\"...\nSuccess! Created Go module \"example.com/foo\" in new directory: foo\n",
			"",
			0,
			&file{"foo", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module example.com/foo\n\ngo 1.18\n"), nil},
				{".gitignore", filePerms, []byte("foo"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
			}},
		},
		{
			[]string{"--nova", "example.com/foo/bar"},
			"Creating Go module \"example.com/foo/bar\"...\nSuccess! Created Go module \"example.com/foo/bar\" in new directory: bar\n",
			"",
			0,
			&file{"bar", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module example.com/foo/bar\n\ngo 1.18\n"), nil},
				{".gitignore", filePerms, []byte("bar"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
		},
		{
			[]string{"-n", "example.com/foo/bar/baz"},
			"Creating Go module \"example.com/foo/bar/baz\"...\nSuccess! Created Go module \"example.com/foo/bar/baz\" in new directory: baz\n",
			"",
			0,
			&file{"baz", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module example.com/foo/bar/baz\n\ngo 1.18\n"), nil},
				{".gitignore", filePerms, []byte("baz"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
		},
	}

	testDir := setUpTestDir(t)
	defer tearDownTestDir(t, testDir)

	for _, tc := range tests {
		input := fmt.Sprintf("%s", tc.args)

		var outputBuffer bytes.Buffer
		var errorOutputBuffer bytes.Buffer
		exitCodeHandler := func(exitCode int) {
			// Test: Exit code
			if tc.expectedExitCode != exitCode {
				t.Errorf(testCaseUnexpectedMessage(input, "exit code", tc.expectedExitCode, exitCode))
			}
		}
		app := cli.AppWithCustomOutputAndExit(&outputBuffer, &errorOutputBuffer, exitCodeHandler)
		_ = app.Run(append([]string{cli.Name}, tc.args...))
		actualOutput := outputBuffer.String()
		actualErrorOutput := errorOutputBuffer.String()

		// Test: Output
		if actualOutput != tc.expectedOutput {
			t.Error(testCaseUnexpectedMessage(input, "output", tc.expectedOutput, actualOutput))
		}

		// Test: Error output
		if actualErrorOutput != tc.expectedErrorOutput {
			t.Error(testCaseUnexpectedMessage(input, "error output", tc.expectedErrorOutput, actualErrorOutput))
		}

		// Test: Files created
		cwd, err := os.Getwd()
		if err != nil {
			t.Error("Could not get cwd", err)
		}
		if tc.expectedFiles != nil {
			walkDir(*tc.expectedFiles, cwd, func(f file, root string) {
				filePath := path.Join(root, f.name)
				assertExpectedFileIsAtPath(t, input, f, filePath)
			})
		} else {
			actualEntries, err := os.ReadDir(cwd)
			if err != nil {
				errorMessage := fmt.Sprintf("Unable to read current directory: %s", cwd)
				t.Error(testInputUnexpectedMessage(input, errorMessage))
			} else {
				if len(actualEntries) > 0 {
					fileNames := []string{}
					for _, actualEntry := range actualEntries {
						fileNames = append(fileNames, actualEntry.Name())
					}
					errorMessage := fmt.Sprintf("Files were created when none were expected: %v", fileNames)
					t.Error(testInputUnexpectedMessage(input, errorMessage))
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

func testCaseUnexpectedMessage[T any](input string, topic string, expected T, actual T) string {
	testResultMessage := testResultUnexpectedMessage(topic, expected, actual)
	testCaseMessage := testInputUnexpectedMessage(input, testResultMessage)
	return testCaseMessage
}

func testInputUnexpectedMessage(input string, message string) string {
	return fmt.Sprintf("Test with input %s: %s", input, message)
}

func testResultUnexpectedMessage[T any](thing string, expected T, actual T) string {
	return fmt.Sprintf("Unexpected %s\nExpected: %v\nActual  : %v\n", thing, expected, actual)
}

func walkDir(f file, root string, fn func(file, string)) {
	fn(f, root)

	if f.files != nil {
		root = path.Join(root, f.name)
		for _, childFile := range f.files {
			walkDir(childFile, root, fn)
		}
	}
}

func assertExpectedFileIsAtPath(t *testing.T, input string, f file, filePath string) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		errorMessage := fmt.Sprintf("Unable to stat expected file: %s", filePath)
		t.Error(testInputUnexpectedMessage(input, errorMessage))
		return
	}

	actualMode := fileInfo.Mode()
	if f.perm != actualMode {
		errorMessage := testResultUnexpectedMessage(fmt.Sprintf("file perms at path: %s", filePath), f.perm, actualMode)
		t.Error(errorMessage)
	}

	if f.content != nil {
		// Compare files
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			errorMessage := fmt.Sprintf("Unable to read expected file: %s", filePath)
			t.Error(testInputUnexpectedMessage(input, errorMessage))
		} else {
			expectedFileContent := string(f.content)
			actualFileContent := string(bytes)
			if expectedFileContent != actualFileContent {
				errorMessage := testResultUnexpectedMessage(fmt.Sprintf("file content at path: %s", filePath), expectedFileContent, actualFileContent)
				t.Error(testInputUnexpectedMessage(input, errorMessage))
			}
		}
	} else {
		// Compare dirs
		actualEntries, err := os.ReadDir(filePath)
		if err != nil {
			errorMessage := fmt.Sprintf("Unable to read expected directory: %s", filePath)
			t.Error(errorMessage)
		} else {
			expectedEntriesExist := map[string]bool{}
			for _, expectedEntry := range f.files {
				expectedEntriesExist[expectedEntry.name] = false
			}

			for _, actualEntry := range actualEntries {
				actualFileName := actualEntry.Name()
				_, ok := expectedEntriesExist[actualFileName]
				if !ok {
					errorMessage := fmt.Sprintf("Unexpected file exists: %s", path.Join(filePath, actualFileName))
					t.Error(testInputUnexpectedMessage(input, errorMessage))
				} else {
					expectedEntriesExist[actualFileName] = true
				}
			}

			for fileName, wasFound := range expectedEntriesExist {
				if !wasFound {
					errorMessage := fmt.Sprintf("Expected file not found: %s", path.Join(filePath, fileName))
					t.Error(testInputUnexpectedMessage(input, errorMessage))
				}
			}
		}
	}
}
