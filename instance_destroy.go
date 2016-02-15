package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
)

var (
	cmdDestroyInstance = &cobra.Command{
		Short: "Destroy a single instance",
		Long:  "Destroy a single instance",
		Use:   "destroy",
		Run:   destroyInstance,
	}

	destroyInstanceFlags providers.ClusterInstanceInfo
)

func init() {
	cmdDestroyInstance.Flags().StringVar(&destroyInstanceFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdDestroyInstance.Flags().StringVar(&destroyInstanceFlags.Name, "name", "", "Cluster name")
	cmdDestroyInstance.Flags().StringVar(&destroyInstanceFlags.Prefix, "prefix", "", "Instance prefix name")
	cmdInstance.AddCommand(cmdDestroyInstance)
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

	if err := provider.DeleteInstance(destroyInstanceFlags, newDnsProvider()); err != nil {
		Exitf("Failed to destroy instance: %v\n", err)
	}

	// Update existing members
	isEtcdProxy := func(i providers.ClusterInstance) bool {
		return false
	}
	if err := providers.UpdateClusterMembers(log, destroyInstanceFlags.ClusterInfo, isEtcdProxy, provider); err != nil {
		Exitf("Failed to update cluster members: %v\n", err)
	}

	Infof("Destroyed instance %s\n", destroyInstanceFlags)
}
