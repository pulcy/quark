package main

import (
	"github.com/spf13/cobra"
)

var (
	cmdPlans = &cobra.Command{
		Use: "plans",
		Run: showPlans,
	}
)

func init() {
	cmdMain.AddCommand(cmdPlans)
}

func showPlans(cmd *cobra.Command, args []string) {
	provider := newProvider()
	err := provider.ShowPlans()
	if err != nil {
		Exitf("Failed to show plans: %v\n", err)
	}
}
