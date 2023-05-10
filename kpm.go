// Copyright 2022 The KCL Authors. All rights reserved.

package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"kusionstack.io/kpm/pkg/cmd"
	"kusionstack.io/kpm/pkg/reporter"
	"kusionstack.io/kpm/pkg/settings"
)

var version string

func main() {
	reporter.InitReporter()
	setting, err := settings.Init()
	if err != nil {
		reporter.Fatal(err)
	}
	app := cli.NewApp()
	app.Name = "kpm"
	app.Usage = "kpm is a kcl package manager"
	app.Version = version
	app.UsageText = "kpm  <command> [arguments]..."
	app.Commands = []*cli.Command{
		cmd.NewInitCmd(),
		cmd.NewAddCmd(),
		cmd.NewPkgCmd(),
		cmd.NewMetadataCmd(),
		cmd.NewRunCmd(setting),
		cmd.NewLoginCmd(setting),
		cmd.NewLogoutCmd(setting),
		cmd.NewPushCmd(setting),
		cmd.NewPullCmd(setting),
	}
	err = app.Run(os.Args)
	if err != nil {
		reporter.Fatal(err)
	}
}
