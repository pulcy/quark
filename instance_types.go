package main

import (
	"github.com/spf13/cobra"
)

var (
	cmdInstanceTypes = &cobra.Command{
		Use: "types",
		Run: showInstanceTypes,
	}
)

func init() {
	cmdInstance.AddCommand(cmdInstanceTypes)
}

func showInstanceTypes(cmd *cobra.Command, args []string) {
	provider := newProvider()
	err := provider.ShowInstanceTypes()
	if err != nil {
		Exitf("Failed to show instance types: %v\n", err)
	}
}
