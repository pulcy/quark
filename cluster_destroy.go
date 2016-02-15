package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
)

var (
	cmdDestroyCluster = &cobra.Command{
		Short: "Destroy an entire cluster",
		Long:  "Destroy an entire cluster",
		Use:   "destroy",
		Run:   destroyCluster,
	}

	destroyClusterFlags providers.ClusterInfo
)

func init() {
	cmdDestroyCluster.Flags().StringVar(&destroyClusterFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdDestroyCluster.Flags().StringVar(&destroyClusterFlags.Name, "name", "", "Cluster name")
	cmdCluster.AddCommand(cmdDestroyCluster)
}

func destroyCluster(cmd *cobra.Command, args []string) {
	clusterInfoFromArgs(&destroyClusterFlags, args)

	if destroyClusterFlags.Domain == "" {
		Exitf("Please specify a domain\n")
	}
	if destroyClusterFlags.Name == "" {
		Exitf("Please specify a name\n")
	}
	if err := confirm(fmt.Sprintf("Are you sure you want to destroy %s?", destroyClusterFlags.String())); err != nil {
		Exitf("%v\n", err)
	}
	provider := newProvider()
	err := provider.DeleteCluster(destroyClusterFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to destroy cluster: %v\n", err)
	}
}
