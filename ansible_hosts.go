package main

import (
	"github.com/spf13/cobra"
)

var (
	cmdAnsibleHosts = &cobra.Command{
		Use: "ahosts",
		Run: ansibleHosts,
	}

	sshPort        int
	developersJson string
)

func init() {
	cmdAnsibleHosts.Flags().StringVar(&developersJson, "developers", "", "Path of developers office file")
	cmdAnsibleHosts.Flags().IntVar(&sshPort, "ssh-port", 0, "SSH port used fo ansible access")
	cmdMain.AddCommand(cmdAnsibleHosts)
}

func ansibleHosts(cmd *cobra.Command, args []string) {
	if developersJson == "" {
		Exitf("Please specific a developers json file path\n")
	}
	provider := newProvider()
	err := provider.CreateAnsibleHosts(domain, sshPort, developersJson)
	if err != nil {
		Exitf("Failed to create ansible-hosts: %v\n", err)
	}
}
