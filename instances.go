package main

import (
	"fmt"

	"arvika.pulcy.com/iggi/droplets/providers"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

var (
	cmdInstances = &cobra.Command{
		Use: "instances",
		Run: showInstances,
	}

	instancesFlags providers.ClusterInfo
)

func init() {
	cmdInstances.Flags().StringVar(&instancesFlags.Domain, "domain", defaultDomain, "Cluster domain")
	cmdInstances.Flags().StringVar(&instancesFlags.Name, "name", "", "Cluster name")
	cmdMain.AddCommand(cmdInstances)
}

func showInstances(cmd *cobra.Command, args []string) {
	if instancesFlags.Name == "" {
		Exitf("Please specify a name\n")
	}
	provider := newProvider()
	instances, err := provider.GetInstances(&instancesFlags)
	if err != nil {
		Exitf("Failed to list instances: %v\n", err)
	}
	lines := []string{"Name | Private IP | Public IP"}
	for _, i := range instances {
		lines = append(lines, fmt.Sprintf("%s | %s | %s", i.Name, i.PrivateIpv4, i.PublicIpv4))
	}
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)
}
