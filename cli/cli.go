package cli

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
)

const Version string = "v0.0.1"

func App() *cli.App {
	return &cli.App{
		Name:            "go-mod-create",
		Usage:           "Creates a new Go module",
		Description:     "https://github.com/jbrudvik/go-mod-create",
		Version:         Version,
		HideHelpCommand: true,
		ArgsUsage:       "[module]",
		Action: func(c *cli.Context) error {
			args := c.Args()
			if args.Len() != 1 {
				cli.ShowAppHelpAndExit(c, 0)
				return nil
			} else {
				module := args.First()
				fmt.Printf("Creating module \"%s\"...\n", module)
				err := CreateModule(module)
				if err != nil {
					errorMessage := fmt.Sprintf("Failed to create module \"%s\": %s", module, err)
					return cli.Exit(errorMessage, 1)
				}
			}
			return nil
		},
	}
}

// TODO: Factor this out to a separate module -- so we can test it well
/*
- If the directory already exists, don't touch it. Do throw an error.
*/
func CreateModule(module string) error {
	// TODO: Implement
	return errors.New("uh oh")
}
