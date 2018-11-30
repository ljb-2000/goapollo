package main

import (
	"github.com/lifei6671/goapollo/cmd"
	"gopkg.in/urfave/cli.v2"
	"os"
	"github.com/sirupsen/logrus"
	"github.com/lifei6671/goapollo/log"
)

const APP_VERSION = "0.1"

//go run main.go run -app_id=6e77bd897fe903ac -ns=TEST1.nginx -server_url=http://dev.config.apollo.com/
func main()  {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:      true,
		QuoteEmptyFields: true,
		FullTimestamp:    true,
	})
	logrus.AddHook(&log.ContextHook{ LogPath: "./runtime/logs/"})

	app := &cli.App{}
	app.Name = "Apollo client"
	app.Usage = "A Apollo client"
	app.Version = APP_VERSION
	app.Commands = []*cli.Command{
		cmd.Start,
	}
	app.Run(os.Args)
}
