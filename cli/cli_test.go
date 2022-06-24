package cli_test

import (
	"bytes"
	"fmt"
	"strconv"
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

var versionOutput string = fmt.Sprintf("gmc version %s\n", cli.Version)

const errorMessageUnknownFlag string = "Error: Unknown flag\n\n"
const errorMessageModuleNameRequired string = "Error: Module name is required\n\n"
const errorMessageTooManyModuleNames string = "Error: Only one module name is allowed\n\n"

// TODO: If I need my test helpers, consider just open-sourcing them in a separate repo
// - However, make the stack traces work before I do that

// TODO: Consider adding in test title to test cases -- so failures are clearer
// ...or just creating dynamically from the inputs?
// - If I like this approach, do the same with Neat.

func TestRun(t *testing.T) {
	tests := []struct {
		args                []string
		expectedOutput      string
		expectedErrorOutput string
		expectedExitCode    int
	}{
		{
			[]string{"-h"},
			helpOutput,
			"",
			0,
		},
		{
			[]string{"--help"},
			helpOutput,
			"",
			0,
		},
		{
			[]string{"-v"},
			versionOutput,
			"",
			0,
		},
		{
			[]string{"--vesion"},
			versionOutput,
			"",
			0,
		},
		{
			[]string{},
			helpOutput,
			errorMessageModuleNameRequired,
			1,
		},
		{
			[]string{"-e"},
			helpOutput,
			errorMessageUnknownFlag,
			1,
		},
		{
			[]string{"-n"},
			helpOutput,
			errorMessageModuleNameRequired,
			1,
		},
		{
			[]string{"--nova"},
			helpOutput,
			errorMessageModuleNameRequired,
			1,
		},
		{
			[]string{"a1", "a2"},
			helpOutput,
			errorMessageTooManyModuleNames,
			1,
		},
		// TODO: Add
		// - a1
		// - --nova a1
		// - example.com/foo/bar
		// - --nova example.com/foo/bar
	}

	for _, tc := range tests {
		input := fmt.Sprintf("%s", tc.args)

		var outputBuffer bytes.Buffer
		var errorOutputBuffer bytes.Buffer
		exitCodeHandler := func(exitCode int) {
			if tc.expectedExitCode != exitCode {
				t.Errorf(testCaseFailureMessage(input, "exit code", strconv.Itoa(tc.expectedExitCode), strconv.Itoa(exitCode)))
			}
		}

		app := cli.AppWithCustomOutputAndExit(&outputBuffer, &errorOutputBuffer, exitCodeHandler)
		_ = app.Run(append([]string{"gmc"}, tc.args...))

		actualOutput := outputBuffer.String()
		actualErrorOutput := errorOutputBuffer.String()
		if actualOutput != tc.expectedOutput {
			t.Error(testCaseFailureMessage(input, "output", tc.expectedOutput, actualOutput))
		}
		if actualErrorOutput != tc.expectedErrorOutput {
			t.Error(testCaseFailureMessage(input, "error output", tc.expectedErrorOutput, actualErrorOutput))
		}
	}
}

// TODO: Rename `what` parameter
func testCaseFailureMessage(input string, what string, expected string, actual string) string {
	return fmt.Sprintf("Test with input %s: Unexpected error output\nExpected: %s\nActual  : %s\n", input, expected, actual)
}
