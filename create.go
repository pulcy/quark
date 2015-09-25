package main

import (
	"arvika.pulcy.com/iggi/droplets/providers"
	"github.com/spf13/cobra"
)

const (
	defaultDomain        = "iggi.xyz"
	defaultClusterImage  = "coreos-stable"
	defaultClusterRegion = "ams3"
	defaultClusterSize   = "512mb"
	defaultInstanceCount = 3
	sshKey               = "ewout@prangsma.net"
)

var (
	cmdCreate = &cobra.Command{
		Use: "create",
		Run: showUsage,
	}
	cmdCreateCluster = &cobra.Command{
		Use: "cluster",
		Run: createCluster,
	}

	createClusterFlags providers.CreateClusterOptions
)

func init() {
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Domain, "domain", defaultDomain, "Cluster domain")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Name, "name", "", "Cluster name")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Image, "image", defaultClusterImage, "OS image to run on new droplets")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Region, "region", defaultClusterRegion, "Region to create the droplets in")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Size, "size", defaultClusterSize, "Size of the new droplet")
	cmdCreateCluster.Flags().IntVar(&createClusterFlags.InstanceCount, "instance-count", defaultInstanceCount, "Number of instances in cluster")
	cmdCreate.AddCommand(cmdCreateCluster)
	cmdMain.AddCommand(cmdCreate)
}

func createCluster(cmd *cobra.Command, args []string) {
	createClusterFlags.SSHKeyNames = []string{sshKey}
	provider := newProvider()

	// Validate
	if err := createClusterFlags.Validate(); err != nil {
		Exitf("%s\n", err.Error())
	}

	// See if there are already instances for the given cluster
	instances, err := provider.GetInstances(&createClusterFlags.ClusterInfo)
	if err != nil {
		Exitf("Failed to query existing instances: %v\n", err)
	}
	if len(instances) > 0 {
		Exitf("Cluster %s.%s already exists.\n", createClusterFlags.Name, createClusterFlags.Domain)
	}

	// Create
	err = provider.CreateCluster(&createClusterFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to create new cluster: %v\n", err)
	}
}
