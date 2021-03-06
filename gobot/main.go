package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/potix/gobot"
)

func main() {
	app := cli.NewApp()
	app.Name = "gobot"
	app.Author = "The Gobot team"
	app.Email = "https://github.com/potix/gobot"
	app.Version = gobot.Version()
	app.Usage = "Command Line Utility for Gobot"
	app.Commands = []cli.Command{
		Generate(),
	}
	app.Run(os.Args)
}
