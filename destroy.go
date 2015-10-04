package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"arvika.pulcy.com/pulcy/droplets/providers"
)

var (
	cmdDestroy = &cobra.Command{
		Use: "destroy",
		Run: showUsage,
	}
	cmdDestroyCluster = &cobra.Command{
		Short: "Destroy an entire cluster",
		Long:  "Destroy an entire cluster",
		Use:   "cluster",
		Run:   destroyCluster,
	}

	destroyClusterFlags providers.ClusterInfo
)

func init() {
	cmdDestroyCluster.Flags().StringVar(&destroyClusterFlags.Domain, "domain", def("DROPLETS_DOMAIN", defaultDomain), "Cluster domain")
	cmdDestroyCluster.Flags().StringVar(&destroyClusterFlags.Name, "name", "", "Cluster name")
	cmdDestroy.AddCommand(cmdDestroyCluster)
	cmdMain.AddCommand(cmdDestroy)
}

func destroyCluster(cmd *cobra.Command, args []string) {
	if destroyClusterFlags.Domain == "" {
		Exitf("Please specify a domain\n")
	}
	if destroyClusterFlags.Name == "" {
		Exitf("Please specify a name\n")
	}
	if err := confirm(fmt.Sprintf("Are you sure you want to destroy %s.%s?", destroyClusterFlags.Name, destroyClusterFlags.Domain)); err != nil {
		Exitf("%v\n", err)
	}
	provider := newProvider()
	err := provider.DeleteCluster(&destroyClusterFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to destroy cluster: %v\n", err)
	}
}
