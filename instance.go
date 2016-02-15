package main

import (
	"github.com/spf13/cobra"
)

var (
	cmdInstance = &cobra.Command{
		Use: "instance",
		Run: showUsage,
	}
)

func init() {
	cmdMain.AddCommand(cmdInstance)
}
