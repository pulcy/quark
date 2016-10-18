// Copyright (C) 2015 Scaleway. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.md file.

package cli

import "github.com/scaleway/scaleway-cli/pkg/commands"

var cmdLogs = &Command{
	Exec:        runLogs,
	UsageLine:   "logs [OPTIONS] SERVER",
	Description: "Fetch the logs of a server",
	Help:        "Fetch the logs of a server.",
}

func init() {
	cmdLogs.Flag.BoolVar(&logsHelp, []string{"h", "-help"}, false, "Print usage")
	cmdLogs.Flag.StringVar(&logsGateway, []string{"g", "-gateway"}, "", "Use a SSH gateway")
}

// FLags
var logsHelp bool      // -h, --help flag
var logsGateway string // -g, --gateway flag

func runLogs(cmd *Command, rawArgs []string) error {
	if logsHelp {
		return cmd.PrintUsage()
	}
	if len(rawArgs) != 1 {
		return cmd.PrintShortUsage()
	}

	args := commands.LogsArgs{
		Gateway: logsGateway,
		Server:  rawArgs[0],
	}
	ctx := cmd.GetContext(rawArgs)
	return commands.RunLogs(ctx, args)
}
