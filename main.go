package main

import (
	"log"
	"os"

	"github.com/jbrudvik/gmc/cli"
)

func main() {
	app := cli.App()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
