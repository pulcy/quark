package main

import (
	"github.com/spf13/cobra"
)

var (
	cmdDns = &cobra.Command{
		Use: "dns",
		Run: showUsage,
	}
	cmdDnsRecords = &cobra.Command{
		Use: "records",
		Run: showDnsRecords,
	}

	dnsFlags struct {
		Domain string
	}
)

func init() {
	cmdDnsRecords.Flags().StringVar(&dnsFlags.Domain, "domain", defaultDomain(), "Domain name")
	cmdDns.AddCommand(cmdDnsRecords)
	cmdMain.AddCommand(cmdDns)
}

func showDnsRecords(cmd *cobra.Command, args []string) {
	if dnsFlags.Domain == "" {
		Exitf("Please specify a domain\n")
	}
	provider := newDnsProvider()
	err := provider.ShowDomainRecords(dnsFlags.Domain)
	if err != nil {
		Exitf("Failed to show dns records: %v\n", err)
	}
}
