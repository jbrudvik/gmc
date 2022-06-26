package cli

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/urfave/cli/v2"
)

const Name string = "gmc"
const Version string = "v0.0.1"
const Url string = "https://github.com/jbrudvik/" + Name
const Description string = Name + " [module name] creates a directory containing:\n" +
	"- go.mod            Go module metadata\n" +
	"- .gitignore        ignores your module's binary\n" +
	"- main.go           your module's first code\n" +
	"- .nova (Optional)  Nova editor configuration\n" +
	"\n" +
	"This directory can be immediately built/run/installed using the `go` CLI.\n" +
	"\n" +
	"More information: " + Url

//go:embed all:assets
var assets embed.FS

const assetsDefaultDir string = "default"

func App() *cli.App {
	return AppWithCustomOutput(os.Stdout, os.Stderr)
}

func AppWithCustomOutput(output io.Writer, errorOutput io.Writer) *cli.App {
	exitCodeHandler := func(exitCode int) {
		os.Exit(exitCode)
	}
	return AppWithCustomOutputAndExit(os.Stdout, os.Stderr, exitCodeHandler)
}

func AppWithCustomOutputAndExit(output io.Writer, errorOutput io.Writer, exitCodeHandler func(int)) *cli.App {
	return &cli.App{
		Writer:    output,
		ErrWriter: errorOutput,
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				errorMessage := fmt.Sprintf("%s\n\n", err)
				fmt.Fprint(c.App.ErrWriter, errorMessage)
				cli.ShowAppHelp(c)
				exitCodeHandler(1)
			} else {
				exitCodeHandler(0)
			}
		},
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			return errors.New("Error: Unknown flag")
		},
		Name:            Name,
		Usage:           "(Go mod create) creates Go modules",
		Description:     Description,
		Version:         Version,
		HideHelpCommand: true,
		ArgsUsage:       "[module name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "nova",
				Usage:   "include Nova configuration",
				Aliases: []string{"n"},
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args()
			if args.Len() < 1 {
				errorMessage := fmt.Sprintf("Error: Module name is required")
				return errors.New(errorMessage)
			} else if args.Len() > 1 {
				// TODO: Control showing help or not? Or just do it all the time?
				errorMessage := fmt.Sprintf("Error: Only one module name is allowed")
				return errors.New(errorMessage)
			} else {
				module := args.First()
				var extraDirs []string
				if c.Bool("nova") {
					extraDirs = append(extraDirs, "nova")
				}

				fmt.Fprintf(output, "Creating Go module \"%s\"...\n", module)
				moduleDir, err := createModule(module, extraDirs)

				if err != nil {
					errorMessage := fmt.Sprintf("Failed to create Go module \"%s\": %s", module, err)
					return errors.New(errorMessage)
				} else {
					successMessage := fmt.Sprintf("Success! Created Go module \"%s\"", module)
					if moduleDir != nil {
						// NOTE: This should always execute
						successMessage += fmt.Sprintf(" in new directory: %s", *moduleDir)
					}
					fmt.Fprintf(output, "%s\n", successMessage)
				}
			}
			return nil
		},
	}
}

func createModule(module string, extraDirs []string) (*string, error) {
	moduleDir := path.Base(module)

	// Create module directory && change into the directory
	err := os.Mkdir(moduleDir, 0755)
	if err != nil {
		return nil, err
	}
	err = os.Chdir(moduleDir)
	if err != nil {
		return nil, err
	}
	defer os.Chdir("..")

	// Create go.mod
	cmd := exec.Command("go", "mod", "init", module)
	if err = cmd.Run(); err != nil {
		return nil, err
	}

	// Create .gitignore
	err = os.WriteFile(".gitignore", []byte(moduleDir), 0644)
	if err != nil {
		return nil, err
	}

	// Copy over assets
	err = copyEmbeddedFS(assets, assetsDefaultDir)
	if err != nil {
		return nil, err
	}

	// Copy over extras
	for _, extraDir := range extraDirs {
		err = copyEmbeddedFS(assets, extraDir)
		if err != nil {
			return nil, err
		}
	}

	return &moduleDir, nil
}

func copyEmbeddedFS(srcFS embed.FS, src string) error {
	entries, err := srcFS.ReadDir(".")
	if err != nil {
		return err
	}
	dir := entries[0] // The `assets` dir // TODO: Q: Should this be a constant?
	root := dir.Name()
	root = path.Join(root, src)

	err = fs.WalkDir(srcFS, root, func(srcPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// TODO: This really can't be right. Right? We should just be walking a level lower... right?
		// TODO: This more slickly
		dstPath := strings.TrimPrefix(srcPath, root)
		if dstPath == "" {
			return nil
		}
		dstPath = strings.TrimPrefix(dstPath, "/")

		if entry.IsDir() {
			// Create dir
			err = os.Mkdir(path.Join(".", dstPath), 0755)
			if err != nil {
				return err
			}
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
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
