package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
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
	cmdDestroyInstance = &cobra.Command{
		Short: "Destroy a single instance",
		Long:  "Destroy a single instance",
		Use:   "instance",
		Run:   destroyInstance,
	}

	destroyClusterFlags  providers.ClusterInfo
	destroyInstanceFlags providers.ClusterInstanceInfo
)

func init() {
	cmdDestroyCluster.Flags().StringVar(&destroyClusterFlags.Domain, "domain", def("QUARK_DOMAIN", "domain", defaultDomain), "Cluster domain")
	cmdDestroyCluster.Flags().StringVar(&destroyClusterFlags.Name, "name", "", "Cluster name")
	cmdDestroy.AddCommand(cmdDestroyCluster)

	cmdDestroyInstance.Flags().StringVar(&destroyInstanceFlags.Domain, "domain", def("QUARK_DOMAIN", "domain", defaultDomain), "Cluster domain")
	cmdDestroyInstance.Flags().StringVar(&destroyInstanceFlags.Name, "name", "", "Cluster name")
	cmdDestroyInstance.Flags().StringVar(&destroyInstanceFlags.Prefix, "prefix", "", "Instance prefix name")
	cmdDestroy.AddCommand(cmdDestroyInstance)

	cmdMain.AddCommand(cmdDestroy)
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
	err := provider.DeleteCluster(&destroyClusterFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to destroy cluster: %v\n", err)
	}
}

func destroyInstance(cmd *cobra.Command, args []string) {
	clusterInstanceInfoFromArgs(&destroyInstanceFlags, args)

	if destroyInstanceFlags.Domain == "" {
		Exitf("Please specify a domain\n")
	}
	if destroyInstanceFlags.Name == "" {
		Exitf("Please specify a name\n")
	}
	if destroyInstanceFlags.Prefix == "" {
		Exitf("Please specify a prefix\n")
	}
	if err := confirm(fmt.Sprintf("Are you sure you want to destroy %s?", destroyInstanceFlags.String())); err != nil {
		Exitf("%v\n", err)
	}
	provider := newProvider()

	/* TODO Remove instance from etcd
	instances, err := provider.GetInstances(&destroyInstanceFlags.ClusterInfo)
	if err != nil {
		Exitf("Failed to list instances: %v\n", err)
	}
	*/

	if err := provider.DeleteInstance(&destroyInstanceFlags, newDnsProvider()); err != nil {
		Exitf("Failed to destroy instance: %v\n", err)
	}

	// Update existing members
	if err := providers.UpdateClusterMembers(log, destroyInstanceFlags.ClusterInfo, provider); err != nil {
		Exitf("Failed to update cluster members: %v\n", err)
	}

	Infof("Destroyed instance %s\n", destroyInstanceFlags)
}
