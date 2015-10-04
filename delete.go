package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"arvika.pulcy.com/pulcy/droplets/providers"
)

var (
	cmdDelete = &cobra.Command{
		Use: "delete",
		Run: showUsage,
	}
	cmdDeleteCluster = &cobra.Command{
		Use: "cluster",
		Run: deleteCluster,
	}

	deleteClusterFlags providers.ClusterInfo
)

func init() {
	cmdDeleteCluster.Flags().StringVar(&deleteClusterFlags.Domain, "domain", def("DROPLETS_DOMAIN", defaultDomain), "Cluster domain")
	cmdDeleteCluster.Flags().StringVar(&deleteClusterFlags.Name, "name", "", "Cluster name")
	cmdDelete.AddCommand(cmdDeleteCluster)
	cmdMain.AddCommand(cmdDelete)
}

func deleteCluster(cmd *cobra.Command, args []string) {
	if deleteClusterFlags.Domain == "" {
		Exitf("Please specify a domain\n")
	}
	if deleteClusterFlags.Name == "" {
		Exitf("Please specify a name\n")
	}
	if err := confirm(fmt.Sprintf("Are you sure you want to destroy %s.%s?", deleteClusterFlags.Name, deleteClusterFlags.Domain)); err != nil {
		Exitf("%v\n", err)
	}
	provider := newProvider()
	err := provider.DeleteCluster(&deleteClusterFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to delete cluster: %v\n", err)
	}
}
