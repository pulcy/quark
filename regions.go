package main

import (
	"github.com/spf13/cobra"
)

var (
	cmdRegions = &cobra.Command{
		Use: "regions",
		Run: showRegions,
	}
)

func init() {
	cmdMain.AddCommand(cmdRegions)
}

func showRegions(cmd *cobra.Command, args []string) {
	provider := newProvider()
	err := provider.ShowRegions()
	if err != nil {
		Exitf("Failed to show regions: %v\n", err)
	}
}
