package cli_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jbrudvik/gmc/cli"
)

const editor string = "vim"

var helpOutput string = fmt.Sprintf("NAME:\n"+
	"   %s - (Go mod create) creates Go modules so you can start coding ASAP\n"+
	"\n"+
	"USAGE:\n"+
	"   %s [global options] [module name]\n"+
	"\n"+
	"VERSION:\n"+
	"   %s\n"+
	"\n"+
	"DESCRIPTION:\n"+
	"   `%s [module name]` creates a directory containing:\n"+
	"   - Go module metadata: go.mod\n"+
	"   - A place to start writing code: main.go\n"+
	"   \n"+
	"   This module can be immediately run:\n"+
	"   \n"+
	"       $ go run .\n"+
	"       hello, world!\n"+
	"   \n"+
	"   Optionally, the directory can also include:\n"+
	"   - Git repository setup with .gitignore, README.md\n"+
	"   - Nova editor configuration to build/test/run natively\n"+
	"   \n"+
	"   More information: %s\n"+
	"\n"+
	"GLOBAL OPTIONS:\n"+
	"   --git, -g      create as Git repository (default: false)\n"+
	"   --nova, -n     include Nova configuration (default: false)\n"+
	"   --help, -h     show help (default: false)\n"+
	"   --version, -v  print the version (default: false)\n",
	cli.Name,
	cli.Name,
	cli.Version,
	cli.Name,
	cli.Url,
)

var versionOutput string = fmt.Sprintf("%s version %s\n", cli.Name, cli.Version)

const mainGoContents string = "package main\n" +
	"\n" +
	"import (\n" +
	"	\"fmt\"\n" +
	")\n" +
	"\n" +
	"func main() {\n" +
	"	fmt.Println(\"hello, world!\")\n" +
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

type runTestCaseSetUpData struct {
	tempTestDir              string
	originalFilePath         string
	originalEditorEnvVar     string
	originalGitDefaultBranch string
}

type runTestCaseData struct {
	args                []string
	expectedOutput      string
	expectedErrorOutput string
	expectedExitCode    int
	expectedFiles       *file
	expectedGitRepo     *gitRepo
}

type file struct {
	name    string
	perm    fs.FileMode
	content []byte // Non-nil for file
	files   []file // Non-nil for directory
}

const dirPerms fs.FileMode = 0755 | fs.ModeDir
const filePerms fs.FileMode = 0644

type gitRepo struct {
	dir            string
	commitMessages []string
	remote         *string
}

func TestRun(t *testing.T) {
	tests := []runTestCaseData{
		{
			args:                []string{"-h"},
			expectedOutput:      helpOutput,
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args:                []string{"--help"},
			expectedOutput:      helpOutput,
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args:                []string{"-v"},
			expectedOutput:      versionOutput,
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args:                []string{"--version"},
			expectedOutput:      versionOutput,
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args:                []string{},
			expectedOutput:      helpOutput,
			expectedErrorOutput: errorMessageModuleNameRequired,
			expectedExitCode:    1,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args:                []string{"-e"},
			expectedOutput:      helpOutput,
			expectedErrorOutput: errorMessageUnknownFlag,
			expectedExitCode:    1,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args:                []string{"-e", "a1"},
			expectedOutput:      helpOutput,
			expectedErrorOutput: errorMessageUnknownFlag,
			expectedExitCode:    1,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args:                []string{"-n"},
			expectedOutput:      helpOutput,
			expectedErrorOutput: errorMessageModuleNameRequired,
			expectedExitCode:    1,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args:                []string{"--nova"},
			expectedOutput:      helpOutput,
			expectedErrorOutput: errorMessageModuleNameRequired,
			expectedExitCode:    1,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args:                []string{"a1", "a2"},
			expectedOutput:      helpOutput,
			expectedErrorOutput: errorMessageTooManyModuleNames,
			expectedExitCode:    1,
			expectedFiles:       nil,
			expectedGitRepo:     nil,
		},
		{
			args: []string{"a1"},
			expectedOutput: fmt.Sprintf("Creating Go module \"a1\"...\n"+
				"- Created directory: a1\n"+
				"- Initialized Go module\n"+
				"- Created file     : a1/main.go\n"+
				"\n"+
				"Finished creating Go module \"a1\"\n"+
				"\n"+
				"Next steps:\n"+
				"- Start coding: $ %s a1\n", editor),
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"a1", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module a1\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
			}},
			expectedGitRepo: nil,
		},
		{
			args: []string{"-n", "a2"},
			expectedOutput: "Creating Go module \"a2\"...\n" +
				"- Created directory: a2\n" +
				"- Initialized Go module\n" +
				"- Created file     : a2/main.go\n" +
				"- Created directory: a2/.nova\n" +
				"- Created directory: a2/.nova/Tasks\n" +
				"- Created file     : a2/.nova/Tasks/Go.json\n" +
				"\n" +
				"Finished creating Go module \"a2\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Start coding: $ nova a2\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"a2", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module a2\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
			expectedGitRepo: nil,
		},
		{
			args: []string{"--nova", "a3"},
			expectedOutput: "Creating Go module \"a3\"...\n" +
				"- Created directory: a3\n" +
				"- Initialized Go module\n" +
				"- Created file     : a3/main.go\n" +
				"- Created directory: a3/.nova\n" +
				"- Created directory: a3/.nova/Tasks\n" +
				"- Created file     : a3/.nova/Tasks/Go.json\n" +
				"\n" +
				"Finished creating Go module \"a3\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Start coding: $ nova a3\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"a3", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module a3\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
			expectedGitRepo: nil,
		},
		{
			args: []string{"github.com/foo"},
			expectedOutput: fmt.Sprintf("Creating Go module \"github.com/foo\"...\n"+
				"- Created directory: foo\n"+
				"- Initialized Go module\n"+
				"- Created file     : foo/main.go\n"+
				"\n"+
				"Finished creating Go module \"github.com/foo\"\n"+
				"\n"+
				"Next steps:\n"+
				"- Start coding: $ %s foo\n", editor),
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"foo", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module github.com/foo\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
			}},
			expectedGitRepo: nil,
		},
		{
			args: []string{"--nova", "github.com/foo/bar"},
			expectedOutput: "Creating Go module \"github.com/foo/bar\"...\n" +
				"- Created directory: bar\n" +
				"- Initialized Go module\n" +
				"- Created file     : bar/main.go\n" +
				"- Created directory: bar/.nova\n" +
				"- Created directory: bar/.nova/Tasks\n" +
				"- Created file     : bar/.nova/Tasks/Go.json\n" +
				"\n" +
				"Finished creating Go module \"github.com/foo/bar\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Start coding: $ nova bar\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"bar", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module github.com/foo/bar\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
			expectedGitRepo: nil,
		},
		{
			args: []string{"-n", "github.com/foo/bar/baz"},
			expectedOutput: "Creating Go module \"github.com/foo/bar/baz\"...\n" +
				"- Created directory: baz\n" +
				"- Initialized Go module\n" +
				"- Created file     : baz/main.go\n" +
				"- Created directory: baz/.nova\n" +
				"- Created directory: baz/.nova/Tasks\n" +
				"- Created file     : baz/.nova/Tasks/Go.json\n" +
				"\n" +
				"Finished creating Go module \"github.com/foo/bar/baz\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Start coding: $ nova baz\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"baz", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module github.com/foo/bar/baz\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
			expectedGitRepo: nil,
		},
		{
			args: []string{"--git", "github.com/foo/bar"},
			expectedOutput: "Creating Go module \"github.com/foo/bar\"...\n" +
				"- Created directory: bar\n" +
				"- Initialized Go module\n" +
				"- Created file     : bar/main.go\n" +
				"- Initialized Git repository\n" +
				"- Created file     : bar/.gitignore\n" +
				"- Created file     : bar/README.md\n" +
				"- Committed all files to Git repository\n" +
				"- Added remote for Git repository: git@github.com:foo/bar.git\n" +
				"\n" +
				"Finished creating Go module \"github.com/foo/bar\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Create remote Git repository git@github.com:foo/bar.git: https://github.com/new\n" +
				"- Push to remote Git repository: $ git push -u origin main\n" +
				"- Start coding: $ vim bar\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"bar", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module github.com/foo/bar\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".git", dirPerms, nil, nil},
				{".gitignore", filePerms, []byte("bar"), nil},
				{"README.md", filePerms, []byte("# bar\n\n"), nil},
			}},
			expectedGitRepo: &gitRepo{
				"bar",
				[]string{"Initial commit"},
				ptr("git@github.com:foo/bar.git"),
			},
		},
		{
			args: []string{"-g", "github.com/foo/bar"},
			expectedOutput: "Creating Go module \"github.com/foo/bar\"...\n" +
				"- Created directory: bar\n" +
				"- Initialized Go module\n" +
				"- Created file     : bar/main.go\n" +
				"- Initialized Git repository\n" +
				"- Created file     : bar/.gitignore\n" +
				"- Created file     : bar/README.md\n" +
				"- Committed all files to Git repository\n" +
				"- Added remote for Git repository: git@github.com:foo/bar.git\n" +
				"\n" +
				"Finished creating Go module \"github.com/foo/bar\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Create remote Git repository git@github.com:foo/bar.git: https://github.com/new\n" +
				"- Push to remote Git repository: $ git push -u origin main\n" +
				"- Start coding: $ vim bar\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"bar", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module github.com/foo/bar\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".git", dirPerms, nil, nil},
				{".gitignore", filePerms, []byte("bar"), nil},
				{"README.md", filePerms, []byte("# bar\n\n"), nil},
			}},
			expectedGitRepo: &gitRepo{
				"bar",
				[]string{"Initial commit"},
				ptr("git@github.com:foo/bar.git"),
			},
		},
		{
			args: []string{"--git", "--nova", "github.com/foo/bar"},
			expectedOutput: "Creating Go module \"github.com/foo/bar\"...\n" +
				"- Created directory: bar\n" +
				"- Initialized Go module\n" +
				"- Created file     : bar/main.go\n" +
				"- Created directory: bar/.nova\n" +
				"- Created directory: bar/.nova/Tasks\n" +
				"- Created file     : bar/.nova/Tasks/Go.json\n" +
				"- Initialized Git repository\n" +
				"- Created file     : bar/.gitignore\n" +
				"- Created file     : bar/README.md\n" +
				"- Committed all files to Git repository\n" +
				"- Added remote for Git repository: git@github.com:foo/bar.git\n" +
				"\n" +
				"Finished creating Go module \"github.com/foo/bar\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Create remote Git repository git@github.com:foo/bar.git: https://github.com/new\n" +
				"- Push to remote Git repository: $ git push -u origin main\n" +
				"- Start coding: $ nova bar\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"bar", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module github.com/foo/bar\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".git", dirPerms, nil, nil},
				{".gitignore", filePerms, []byte("bar"), nil},
				{"README.md", filePerms, []byte("# bar\n\n"), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
			expectedGitRepo: &gitRepo{
				"bar",
				[]string{"Initial commit"},
				ptr("git@github.com:foo/bar.git"),
			},
		},
		{
			args: []string{"-g", "-n", "github.com/foo/bar"},
			expectedOutput: "Creating Go module \"github.com/foo/bar\"...\n" +
				"- Created directory: bar\n" +
				"- Initialized Go module\n" +
				"- Created file     : bar/main.go\n" +
				"- Created directory: bar/.nova\n" +
				"- Created directory: bar/.nova/Tasks\n" +
				"- Created file     : bar/.nova/Tasks/Go.json\n" +
				"- Initialized Git repository\n" +
				"- Created file     : bar/.gitignore\n" +
				"- Created file     : bar/README.md\n" +
				"- Committed all files to Git repository\n" +
				"- Added remote for Git repository: git@github.com:foo/bar.git\n" +
				"\n" +
				"Finished creating Go module \"github.com/foo/bar\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Create remote Git repository git@github.com:foo/bar.git: https://github.com/new\n" +
				"- Push to remote Git repository: $ git push -u origin main\n" +
				"- Start coding: $ nova bar\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"bar", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module github.com/foo/bar\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".git", dirPerms, nil, nil},
				{".gitignore", filePerms, []byte("bar"), nil},
				{"README.md", filePerms, []byte("# bar\n\n"), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
			expectedGitRepo: &gitRepo{
				"bar",
				[]string{"Initial commit"},
				ptr("git@github.com:foo/bar.git"),
			},
		},
		{
			args: []string{"--git", "--nova", "foo"},
			expectedOutput: "Creating Go module \"foo\"...\n" +
				"- Created directory: foo\n" +
				"- Initialized Go module\n" +
				"- Created file     : foo/main.go\n" +
				"- Created directory: foo/.nova\n" +
				"- Created directory: foo/.nova/Tasks\n" +
				"- Created file     : foo/.nova/Tasks/Go.json\n" +
				"- Initialized Git repository\n" +
				"- Created file     : foo/.gitignore\n" +
				"- Created file     : foo/README.md\n" +
				"- Committed all files to Git repository\n" +
				"- NOTE: Unable to add remote for Git repository\n" +
				"\n" +
				"Finished creating Go module \"foo\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Create remote Git repository\n" +
				"- Push to remote Git repository: $ git push -u origin main\n" +
				"- Start coding: $ nova foo\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"foo", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module foo\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".git", dirPerms, nil, nil},
				{".gitignore", filePerms, []byte("foo"), nil},
				{"README.md", filePerms, []byte("# foo\n\n"), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
			expectedGitRepo: &gitRepo{
				"foo",
				[]string{"Initial commit"},
				nil,
			},
		},
		{
			args: []string{"--git", "--nova", "example.com/foo/bar"},
			expectedOutput: "Creating Go module \"example.com/foo/bar\"...\n" +
				"- Created directory: bar\n" +
				"- Initialized Go module\n" +
				"- Created file     : bar/main.go\n" +
				"- Created directory: bar/.nova\n" +
				"- Created directory: bar/.nova/Tasks\n" +
				"- Created file     : bar/.nova/Tasks/Go.json\n" +
				"- Initialized Git repository\n" +
				"- Created file     : bar/.gitignore\n" +
				"- Created file     : bar/README.md\n" +
				"- Committed all files to Git repository\n" +
				"- Added remote for Git repository: git@example.com:foo/bar.git\n" +
				"\n" +
				"Finished creating Go module \"example.com/foo/bar\"\n" +
				"\n" +
				"Next steps:\n" +
				"- Create remote Git repository git@example.com:foo/bar.git\n" +
				"- Push to remote Git repository: $ git push -u origin main\n" +
				"- Start coding: $ nova bar\n",
			expectedErrorOutput: "",
			expectedExitCode:    0,
			expectedFiles: &file{"bar", dirPerms, nil, []file{
				{"go.mod", filePerms, []byte("module example.com/foo/bar\n\ngo 1.18\n"), nil},
				{"main.go", filePerms, []byte(mainGoContents), nil},
				{".git", dirPerms, nil, nil},
				{".gitignore", filePerms, []byte("bar"), nil},
				{"README.md", filePerms, []byte("# bar\n\n"), nil},
				{".nova", dirPerms, nil, []file{
					{"Tasks", dirPerms, nil, []file{
						{"Go.json", filePerms, []byte(novaTaskContents), nil},
					}},
				}},
			}},
			expectedGitRepo: &gitRepo{
				"bar",
				[]string{"Initial commit"},
				ptr("git@example.com:foo/bar.git"),
			},
		},
	}

	for _, tc := range tests {
		testName := strings.Join(tc.args, " ")
		t.Run(testName, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

func runTestCase(t *testing.T, tc runTestCaseData) {
	setUpData := runTestCaseSetUp(t)
	defer runTestCaseTearDown(t, setUpData)

	var outputBuffer bytes.Buffer
	var errorOutputBuffer bytes.Buffer
	exitCodeHandler := func(exitCode int) {
		// Test: Exit code
		if tc.expectedExitCode != exitCode {
			t.Errorf(testCaseUnexpectedMessage("exit code", tc.expectedExitCode, exitCode))
		}
	}
	app := cli.AppWithCustomOutputAndExit(&outputBuffer, &errorOutputBuffer, exitCodeHandler)
	args := append([]string{cli.Name}, tc.args...)
	_ = app.Run(args)
	actualOutput := outputBuffer.String()
	actualErrorOutput := errorOutputBuffer.String()

	// Test: Output
	if actualOutput != tc.expectedOutput {
		t.Error(testCaseUnexpectedMessage("output", tc.expectedOutput, actualOutput))
	}

	// Test: Error output
	if actualErrorOutput != tc.expectedErrorOutput {
		t.Error(testCaseUnexpectedMessage("error output", tc.expectedErrorOutput, actualErrorOutput))
	}

	// Test: Files created
	assertExpectedFilesExist(t, tc.expectedFiles)

	// Test: Git
	if tc.expectedGitRepo != nil {
		assertExpectedGitRepoExists(t, *tc.expectedGitRepo)
	}
}

func runTestCaseSetUp(t *testing.T) runTestCaseSetUpData {
	errorMessage := "Failure during setup"

	// Get current working directory
	originalFilePath, err := os.Getwd()
	if err != nil {
		t.Fatalf("%s: Failed to get current working directory: %s", errorMessage, err)
	}

	// Create temp test directory
	tempTestDir, err := os.MkdirTemp(".", "cli_test_dir")
	if err != nil {
		t.Fatalf("%s: Failed to create test directory: %s", errorMessage, err)
	}
	err = os.Chdir(tempTestDir)
	if err != nil {
		t.Fatalf("%s: Failed to change into test directory: %s", errorMessage, err)
	}

	// Set EDITOR env var
	originalEditorEnvVar := os.Getenv("EDITOR")
	err = os.Setenv("EDITOR", editor)
	if err != nil {
		t.Fatalf("%s: Failed to set EDITOR env var: %s", errorMessage, err)
	}

	// Set Git default branch
	var cmdOutputBuffer bytes.Buffer
	cmd := exec.Command("git", "config", "--global", "init.defaultBranch")
	cmd.Stdout = &cmdOutputBuffer
	originalGitDefaultBranch := ""
	if err = cmd.Run(); err != nil {
		// Expected if default branch is not set
	} else {
		originalGitDefaultBranch = strings.TrimSpace(cmdOutputBuffer.String())
	}
	cmd = exec.Command("git", "config", "--global", "init.defaultBranch", "main")
	if err = cmd.Run(); err != nil {
		t.Fatalf("%s: Unable to set Git default branch: %s", errorMessage, err)
	}

	result := runTestCaseSetUpData{
		tempTestDir:              tempTestDir,
		originalFilePath:         originalFilePath,
		originalEditorEnvVar:     originalEditorEnvVar,
		originalGitDefaultBranch: originalGitDefaultBranch,
	}
	return result
}

func runTestCaseTearDown(t *testing.T, setUpData runTestCaseSetUpData) {
	errorMessage := "Failure during teardown"

	// Set Git default branch to original value
	cmd := exec.Command("git", "config", "--global", "init.defaultBranch", setUpData.originalGitDefaultBranch)
	if setUpData.originalGitDefaultBranch == "" {
		cmd = exec.Command("git", "config", "--global", "--unset", "init.defaultBranch")
	}
	if err := cmd.Run(); err != nil {
		t.Errorf("%s: Unable to reset Git default branch setting (%s): %s", errorMessage, setUpData.originalGitDefaultBranch, err)
	}

	// Set EDITOR env var to original value
	err := os.Setenv("EDITOR", setUpData.originalEditorEnvVar)
	if err != nil {
		t.Errorf("%s: Failed to set EDITOR env var to the original value (%s): %s", errorMessage, setUpData.originalEditorEnvVar, err)
	}

	// Move to original directory
	err = os.Chdir(setUpData.originalFilePath)
	if err != nil {
		t.Fatalf("%s: Failed to change to original directory (%s): %s", errorMessage, setUpData.originalFilePath, err)
	}

	// Remove temp test directory
	err = os.RemoveAll(setUpData.tempTestDir)
	if err != nil {
		t.Fatalf("%s: Failed to remove temp test directory (%s): %s", errorMessage, setUpData.tempTestDir, err)
	}
}

func assertExpectedFilesExist(t *testing.T, expectedFiles *file) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error("Could not get cwd", err)
	}
	if expectedFiles != nil {
		walkDir(*expectedFiles, cwd, func(f file, root string) {
			filePath := filepath.Join(root, f.name)
			assertExpectedFileIsAtPath(t, f, filePath)
		})
	} else {
		actualEntries, err := os.ReadDir(cwd)
		if err != nil {
			t.Errorf("Unable to read current directory: %s", cwd)
		} else {
			if len(actualEntries) > 0 {
				fileNames := []string{}
				for _, actualEntry := range actualEntries {
					fileNames = append(fileNames, actualEntry.Name())
				}
				t.Errorf("Files were created when none were expected: %v", fileNames)
			}
		}
	}
}

func walkDir(f file, root string, fn func(file, string)) {
	fn(f, root)

	if f.files != nil {
		root = filepath.Join(root, f.name)
		for _, childFile := range f.files {
			walkDir(childFile, root, fn)
		}
	}
}

func assertExpectedFileIsAtPath(t *testing.T, f file, filePath string) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Errorf("Unable to stat expected file: %s", filePath)
		return
	}

	actualMode := fileInfo.Mode()
	if f.perm != actualMode {
		t.Error(testCaseUnexpectedMessage(fmt.Sprintf("file perms at path: %s", filePath), f.perm, actualMode))
	}

	if f.content != nil {
		// Compare files
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Unable to read expected file: %s", filePath)
		} else {
			expectedFileContent := string(f.content)
			actualFileContent := string(bytes)
			if expectedFileContent != actualFileContent {
				t.Error(testCaseUnexpectedMessage(fmt.Sprintf("file content at path: %s", filePath), expectedFileContent, actualFileContent))
			}
		}
	} else {
		// Compare dirs
		actualEntries, err := os.ReadDir(filePath)
		if err != nil {
			t.Errorf("Unable to read expected directory: %s", filePath)
		} else {
			expectedEntriesExist := map[string]bool{}

			if f.files != nil { // nil -> Ignore contents of directory
				for _, expectedEntry := range f.files {
					expectedEntriesExist[expectedEntry.name] = false
				}

				for _, actualEntry := range actualEntries {
					actualFileName := actualEntry.Name()
					_, ok := expectedEntriesExist[actualFileName]
					if !ok {
						t.Errorf(fmt.Sprintf("Unexpected file exists: %s", filepath.Join(filePath, actualFileName)))
					} else {
						expectedEntriesExist[actualFileName] = true
					}
				}

				for fileName, wasFound := range expectedEntriesExist {
					if !wasFound {
						t.Errorf("Expected file not found: %s", filepath.Join(filePath, fileName))
					}
				}
			}
		}
	}
}

func assertExpectedGitRepoExists(t *testing.T, expectedGitRepo gitRepo) {
	// Assert all files have been committed to Git repository
	var cmdOutputBuffer bytes.Buffer
	cmd := exec.Command("git", "status", "-s")
	cmd.Dir = expectedGitRepo.dir
	cmd.Stdout = &cmdOutputBuffer
	if err := cmd.Run(); err != nil {
		t.Errorf("Unable to view Git status in %s:", expectedGitRepo.dir)
		return
	}
	cmdOutput := strings.TrimSpace(cmdOutputBuffer.String())
	if cmdOutput != "" {
		t.Errorf("Not all files committed to Git repository: %s", cmdOutput)
	}

	// Assert Git repository has expected commit history
	cmdOutputBuffer = bytes.Buffer{}
	cmd = exec.Command("git", "log", "--pretty=%s")
	cmd.Dir = expectedGitRepo.dir
	cmd.Stdout = &cmdOutputBuffer
	if err := cmd.Run(); err != nil {
		t.Errorf("Unable to view Git commit history in %s:", expectedGitRepo.dir)
		return
	}
	cmdOutput = strings.TrimSpace(cmdOutputBuffer.String())
	expectedCommitMessagesString := strings.Join(expectedGitRepo.commitMessages, "\n")
	if expectedCommitMessagesString != cmdOutput {
		t.Error(testCaseUnexpectedMessage("Git repository commit message history", expectedCommitMessagesString, cmdOutput))
	}

	// Assert expected Git remote
	cmdOutputBuffer = bytes.Buffer{}
	cmd = exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = expectedGitRepo.dir
	cmd.Stdout = &cmdOutputBuffer
	var actualGitRemote *string = nil
	if err := cmd.Run(); err != nil {
		// Expected when no remote set
	}
	cmdOutput = strings.TrimSpace(cmdOutputBuffer.String())
	if cmdOutput != "" {
		actualGitRemote = &cmdOutput
	}
	if expectedGitRepo.remote == nil {
		if actualGitRemote != nil {
			t.Error(testCaseUnexpectedMessage("Git remote", expectedGitRepo.remote, actualGitRemote))
		}
	} else {
		expectedGitRemote := *expectedGitRepo.remote
		if expectedGitRemote != cmdOutput {
			t.Error(testCaseUnexpectedMessage("Git remote", *expectedGitRepo.remote, *actualGitRemote))
		}
	}
}

func testCaseUnexpectedMessage[T any](thing string, expected T, actual T) string {
	return fmt.Sprintf("Unexpected %s\nExpected: %v\nActual  : %v\n", thing, expected, actual)
}

func ptr[T any](t T) *T {
	return &t
}
