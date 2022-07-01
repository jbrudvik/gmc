package cli

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

const Name string = "gmc"
const Version string = "v0.0.2"
const Url string = "https://github.com/jbrudvik/" + Name

const Description string = "`" + Name + " [module name]` creates a directory containing:\n" +
	"- Go module metadata: go.mod\n" +
	"- A place to start writing code: main.go\n" +
	"\n" +
	"This module can be immediately run:\n" +
	"\n" +
	"    $ go run .\n" +
	"    hello, world!\n" +
	"\n" +
	"Optionally, the directory can also include:\n" +
	"- Git repository setup with .gitignore, README.md\n" +
	"- Nova editor configuration to build/test/run natively\n" +
	"\n" +
	"More information: " + Url

//go:embed all:assets
var assets embed.FS

const assetsDir string = "assets"
const assetsDefaultDir string = "default"

type gitRepo struct {
	initialBranch *string
}

const gitignoreFileName string = ".gitignore"
const readmeFileName string = "README.md"

func App() *cli.App {
	return AppWithCustomOutput(os.Stdout, os.Stderr)
}

func AppWithCustomOutput(output io.Writer, errorOutput io.Writer) *cli.App {
	exitCodeHandler := func(exitCode int) {
		os.Exit(exitCode)
	}
	return AppWithCustomAll(os.Stdout, os.Stderr, exitCodeHandler, nil)
}

func AppWithCustomAll(output io.Writer, errorOutput io.Writer, exitCodeHandler func(int), gitInitialBranch *string) *cli.App {
	return &cli.App{
		Name:        Name,
		Usage:       "(Go mod create) creates Go modules so you can start coding ASAP",
		Version:     Version,
		Description: Description,
		Writer:      output,
		ErrWriter:   errorOutput,
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				errorMessage := fmt.Sprintf("%s\n", err)
				fmt.Fprint(errorOutput, errorMessage)
				if c.IsSet("help") {
					fmt.Fprint(errorOutput, "\n")
					cli.ShowAppHelp(c)
				}
				exitCodeHandler(1)
			} else {
				exitCodeHandler(0)
			}
		},
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			c.Set("help", "true")
			return errors.New("Error: Unknown flag")
		},
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "git",
				Usage:   "create as Git repository",
				Aliases: []string{"g"},
			},
			&cli.BoolFlag{
				Name:    "nova",
				Usage:   "include Nova configuration",
				Aliases: []string{"n"},
			},
		},
		ArgsUsage: "[module name]",
		Action: func(c *cli.Context) error {
			args := c.Args()
			if args.Len() < 1 {
				c.Set("help", "true")
				return errors.New("Error: Module name is required")
			} else if args.Len() > 1 {
				c.Set("help", "true")
				return errors.New("Error: Only one module name is allowed")
			} else {
				// Get only arg: Module name
				module := args.First()

				// Process flags
				var extraDirs []string

				var repo *gitRepo
				if c.Bool("nova") {
					extraDirs = append(extraDirs, "nova")
				}
				if c.Bool("git") {
					repo = &gitRepo{
						initialBranch: gitInitialBranch,
					}
				}

				// Create module
				err := createModule(module, repo, extraDirs, output)
				if err != nil {
					errorMessage := fmt.Sprintf("Failed to create Go module: %s: %s", module, err)
					return errors.New(errorMessage)
				}
			}
			return nil
		},
	}
}

func createModule(module string, repo *gitRepo, extraDirs []string, output io.Writer) error {
	fmt.Fprintf(output, "Creating Go module: %s\n", module)

	moduleBase := filepath.Base(module)
	nextSteps := []string{}

	// Create module directory && change into the directory
	err := os.Mkdir(moduleBase, 0755)
	if err != nil {
		return err
	}
	reportCreatedDir(output, moduleBase)

	// Create go.mod
	cmd := exec.Command("go", "mod", "init", module)
	cmd.Dir = moduleBase
	if err = cmd.Run(); err != nil {
		return err
	}
	fmt.Fprintln(output, "- Initialized Go module")

	// Copy over assets
	err = copyEmbeddedFS(assets, assetsDefaultDir, moduleBase, output)
	if err != nil {
		return err
	}

	// Copy over extras
	for _, extraDir := range extraDirs {
		err = copyEmbeddedFS(assets, extraDir, moduleBase, output)
		if err != nil {
			return err
		}
	}

	if repo != nil {
		// Initialize Git repository
		cmd := exec.Command("git", "init")
		if repo.initialBranch != nil {
			cmd = exec.Command("git", "init", "--initial-branch", *repo.initialBranch)
		}
		cmd.Dir = moduleBase
		if err = cmd.Run(); err != nil {
			return errors.New("Failed to initialize Git repository")
		}
		fmt.Fprintln(output, "- Initialized Git repository")

		// Create .gitignore
		gitignoreFilePath := filepath.Join(moduleBase, gitignoreFileName)
		err = os.WriteFile(gitignoreFilePath, []byte(moduleBase), 0644)
		if err != nil {
			return err
		}
		reportCreatedFile(output, gitignoreFilePath)

		// Create README.md (with title)
		readmeFilePath := filepath.Join(moduleBase, readmeFileName)
		readmeContent := fmt.Sprintf("# %s\n\n", moduleBase)
		err = os.WriteFile(readmeFilePath, []byte(readmeContent), 0644)
		if err != nil {
			return err
		}
		reportCreatedFile(output, readmeFilePath)

		// Commit all files to Git repository
		cmd = exec.Command("git", "add", ".")
		cmd.Dir = moduleBase
		if err = cmd.Run(); err != nil {
			return errors.New("Failed to stage files for Git commit")
		}
		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = moduleBase
		if err = cmd.Run(); err != nil {
			return errors.New("Failed to commit files into Git repository")
		}
		fmt.Fprintln(output, "- Committed all files to Git repository")

		// Add Git repository remote
		gitUrlCore := strings.Replace(module, "/", ":", 1)
		var gitUrl string
		if gitUrlCore != module {
			gitUrl = fmt.Sprintf("git@%s.git", gitUrlCore)
			cmd = exec.Command("git", "remote", "add", "origin", gitUrl)
			cmd.Dir = moduleBase
			if err = cmd.Run(); err != nil {
				return errors.New("Failed to stage files for Git commit")
			}
			fmt.Fprintf(output, "- Added remote for Git repository: %s\n", gitUrl)
		} else {
			fmt.Fprintln(output, "- NOTE: Unable to add remote for Git repository")
		}

		// Add next step: Create remote repository
		nextStepCreateRemote := "Create remote Git repository"
		if len(gitUrl) > 0 {
			nextStepCreateRemote += fmt.Sprintf(" %s", gitUrl)
			if strings.Contains(gitUrl, "github.com") {
				nextStepCreateRemote += ": https://github.com/new"
			}
		}
		nextSteps = append(nextSteps, nextStepCreateRemote)

		// Add next step: Push to remote
		var cmdOutputBuffer bytes.Buffer
		cmd = exec.Command("git", "symbolic-ref", "--short", "HEAD")
		cmd.Dir = moduleBase
		cmd.Stdout = &cmdOutputBuffer
		_ = cmd.Run()
		cmdOutput := strings.TrimSpace(cmdOutputBuffer.String())
		nextStepPush := "Push to remote Git repository: $ git push -u origin "
		if cmdOutput != "" {
			nextStepPush += cmdOutput
		} else {
			nextStepPush += "$(git branch --show-current)"
		}
		nextSteps = append(nextSteps, nextStepPush)
	}

	// Output success
	successMessage := fmt.Sprintf("\nFinished creating Go module: %s", module)
	fmt.Fprintln(output, successMessage)

	// Add next step: Start coding!
	editor := "$EDITOR"
	editorEnvVar := os.Getenv("EDITOR")
	if editorEnvVar != "" {
		editor = editorEnvVar
	}
	for _, extraDir := range extraDirs {
		if "nova" == extraDir {
			editor = "nova"
			break
		}
	}
	nextSteps = append(nextSteps, fmt.Sprintf("Start coding: $ %s %s", editor, moduleBase))

	// Output next steps
	if len(nextSteps) > 0 {
		fmt.Fprintf(output, "\nNext steps:\n")
		for _, nextStep := range nextSteps {
			fmt.Fprintf(output, "- %s\n", nextStep)
		}
	}

	return nil
}

func copyEmbeddedFS(srcFS embed.FS, src string, moduleBase string, output io.Writer) error {
	srcRoot := filepath.Join(assetsDir, src)

	err := fs.WalkDir(srcFS, srcRoot, func(srcPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if srcRoot == srcPath {
			// Ignore the root -- we only want its contents
			return nil
		}

		dstPath := filepath.Join(moduleBase, withoutFilepathPrefix(srcPath, srcRoot))

		if entry.IsDir() {
			// Create dir
			err = os.Mkdir(dstPath, 0755)
			if err != nil {
				return err
			}
			reportCreatedDir(output, dstPath)
		} else {
			// Copy file
			fileBytes, err := fs.ReadFile(srcFS, srcPath)
			if err != nil {
				return err
			}
			err = os.WriteFile(dstPath, fileBytes, 0644)
			if err != nil {
				return err
			}
			reportCreatedFile(output, dstPath)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func reportCreatedAtPath(output io.Writer, fileType string, filePath string) {
	fmt.Fprintln(output, fmt.Sprintf("- Created %-9s: %s", fileType, filePath))
}

func reportCreatedDir(output io.Writer, filePath string) {
	reportCreatedAtPath(output, "directory", filePath)
}

func reportCreatedFile(output io.Writer, filePath string) {
	reportCreatedAtPath(output, "file", filePath)
}

func withoutFilepathPrefix(filePath string, filePathPrefix string) string {
	filePathPrefixWithSeparator := filePathPrefix + string(filepath.Separator)
	return strings.TrimPrefix(filePath, filePathPrefixWithSeparator)
}
