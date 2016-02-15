package main

import (
	"github.com/spf13/cobra"
)

var (
	cmdKeys = &cobra.Command{
		Use: "keys",
		Run: showKeys,
	}
)

func init() {
	cmdInstance.AddCommand(cmdKeys)
}

func showKeys(cmd *cobra.Command, args []string) {
	provider := newProvider()
	err := provider.ShowKeys()
	if err != nil {
		Exitf("Failed to show keys: %v\n", err)
	}
}
