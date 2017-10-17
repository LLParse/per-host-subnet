package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/per-host-subnet/setting"
	"github.com/urfave/cli"
)

var VERSION = "v0.0.0-dev"

func main() {
	app := cli.NewApp()
	app.Name = "per-host-subnet"
	app.Version = VERSION
	app.Usage = "Support per-host-subnet networking"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug, d",
			EnvVar: "RANCHER_DEBUG",
		},
		cli.StringFlag{
			Name:   "metadata-address",
			Value:  setting.DefaultMetadataAddress,
			EnvVar: "RANCHER_METADATA_ADDRESS",
		},
		cli.BoolFlag{
			Name:   "enable-route-update",
			EnvVar: "RANCHER_ENABLE_ROUTE_UPDATE",
		},
		cli.StringFlag{
			Name:   "route-update-provider",
			EnvVar: "RANCHER_ROUTE_UPDATE_PROVIDER",
			Value:  setting.DefaultRouteUpdateProvider,
		},
	}
	app.Action = appMain
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
