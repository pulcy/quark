package main

import (
	"github.com/spf13/cobra"
)

var (
	cmdCluster = &cobra.Command{
		Use: "cluster",
		Run: showUsage,
	}
)

func init() {
	cmdMain.AddCommand(cmdCluster)
}
