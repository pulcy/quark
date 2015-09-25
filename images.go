package main

import (
	"github.com/spf13/cobra"
)

var (
	cmdImages = &cobra.Command{
		Use: "images",
		Run: showImages,
	}
)

func init() {
	cmdMain.AddCommand(cmdImages)
}

func showImages(cmd *cobra.Command, args []string) {
	provider := newProvider()
	err := provider.ShowImages()
	if err != nil {
		Exitf("Failed to show images: %v\n", err)
	}
}
