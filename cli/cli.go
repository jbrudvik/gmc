package cli

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/urfave/cli/v2"
)

const Version string = "v0.0.1"

//go:embed all:assets
var assets embed.FS

const assetsDefaultDir string = "default"

func App() *cli.App {
	return &cli.App{
		Name:  "gmc",
		Usage: "(Go mod create) Creates a Go module starting point",
		Description: "gmc creates a directory containing:\n" +
			"- go.mod            Go module metadata\n" +
			"- .gitignore        ignores your module's binary\n" +
			"- main.go           a starting place for your module's code\n" +
			"- .nova (Optional)  Nova editor configuration\n" +
			"\n" +
			"This new directory can be immediately built/run/installed using the go CLI.\n" +
			"\n" +
			"More information: https://github.com/jbrudvik/gmc\n" +
			"",

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
				cli.ShowAppHelpAndExit(c, 0)
				return nil
			} else if args.Len() > 1 {
				fmt.Fprintf(os.Stderr, "Error: Only one module name is allowed\n\n")
				cli.ShowAppHelpAndExit(c, 1)
				return nil
			} else {
				module := args.First()
				fmt.Printf("Creating module \"%s\"...\n", module)

				// TODO: Add showing nova (and other) options to above? So it's clear
				extraDirs := []string{} // TODO: Use make instead? Google it.
				if c.Bool("nova") {
					extraDirs = append(extraDirs, "nova")
				}

				err := CreateModule(module, extraDirs)

				if err != nil {
					errorMessage := fmt.Sprintf("Failed to create module \"%s\": %s", module, err)
					return cli.Exit(errorMessage, 1)
				}
				fmt.Println("Success!")
			}
			return nil
		},
	}
}

// TODO: Is extras idiomatic? Do it the right way
// TODO: Factor this out to a separate module. Then test it well.
func CreateModule(module string, extraDirs []string) error {
	moduleBase := path.Base(module)

	// Create module directory && change into the directory
	err := os.Mkdir(moduleBase, 0755)
	if err != nil {
		return err
	}
	err = os.Chdir(moduleBase)
	if err != nil {
		return err
	}

	// Create go.mod
	cmd := exec.Command("go", "mod", "init", module)
	if err = cmd.Run(); err != nil {
		return err
	}

	// Create .gitignore
	err = os.WriteFile(".gitignore", []byte(moduleBase), 0644)
	if err != nil {
		return err
	}

	// Copy over assets
	err = copyEmbeddedFS(assets, assetsDefaultDir)
	if err != nil {
		return err
	}

	// Copy over extras
	for _, extraDir := range extraDirs {
		err = copyEmbeddedFS(assets, extraDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Factor out to own package + Unit test
// - Q: Can test code use its own embed? That's probably the right way to do this.
//     - Especially if we can write to an in-memory FS (to watch that its done correctly)
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
