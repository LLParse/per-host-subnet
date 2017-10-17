package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/per-host-subnet/routeupdate"
	"github.com/rancher/per-host-subnet/setting"
	"github.com/urfave/cli"
)

func appMain(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	done := make(chan error)

	m, err := metadata.NewClientAndWait(fmt.Sprintf(setting.MetadataURL, ctx.String("metadata-address")))
	if err != nil {
		return errors.Wrap(err, "Failed to create metadata client")
	}

	_, err = routeupdate.Run(ctx.String("route-update-provider"), m)
	if err != nil {
		return err
	}

	return <-done
}
