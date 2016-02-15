package main

import (
	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
)

const (
	defaultClusterImage   = "coreos-stable"
	defaultClusterSize    = "512mb"
	defaultInstanceCount  = 3
	defaultGluonImage     = "pulcy/gluon:20160214210824"
	defaultRebootStrategy = "etcd-lock"
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
	cmdCreateInstance = &cobra.Command{
		Use: "instance",
		Run: createInstance,
	}

	createClusterFlags  providers.CreateClusterOptions
	createInstanceFlags providers.CreateInstanceOptions
)

func init() {
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Name, "name", "", "Cluster name")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Image, "image", defaultClusterImage, "OS image to run on new instances")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Region, "region", defaultClusterRegion(), "Region to create the instances in")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Size, "size", defaultClusterSize, "Size of the new instances")
	cmdCreateCluster.Flags().IntVar(&createClusterFlags.InstanceCount, "instance-count", defaultInstanceCount, "Number of instances in cluster")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.GluonImage, "gluon-image", defaultGluonImage, "Image containing gluon")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.RebootStrategy, "reboot-strategy", defaultRebootStrategy, "CoreOS reboot strategy")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.PrivateRegistryUrl, "private-registry-url", defaultPrivateRegistryUrl(), "URL of private docker registry")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.PrivateRegistryUserName, "private-registry-username", defaultPrivateRegistryUserName(), "Username for private registry")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.PrivateRegistryPassword, "private-registry-password", defaultPrivateRegistryPassword(), "Password for private registry")
	cmdCreateCluster.Flags().StringSliceVar(&createClusterFlags.SSHKeyNames, "ssh-key", defaultSshKeys(), "Names of SSH keys to add to instances")
	cmdCreate.AddCommand(cmdCreateCluster)

	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Name, "name", "", "Cluster name")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Image, "image", defaultClusterImage, "OS image to run on new instances")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Region, "region", defaultClusterRegion(), "Region to create the instances in")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Size, "size", defaultClusterSize, "Size of the new instances")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.GluonImage, "gluon-image", defaultGluonImage, "Image containing gluon")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.RebootStrategy, "reboot-strategy", defaultRebootStrategy, "CoreOS reboot strategy")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryUrl, "private-registry-url", defaultPrivateRegistryUrl(), "URL of private docker registry")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryUserName, "private-registry-username", defaultPrivateRegistryUserName(), "Username for private registry")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryPassword, "private-registry-password", defaultPrivateRegistryPassword(), "Password for private registry")
	cmdCreateInstance.Flags().StringSliceVar(&createInstanceFlags.SSHKeyNames, "ssh-key", defaultSshKeys(), "Names of SSH keys to add to instance")
	cmdCreateInstance.Flags().BoolVar(&createInstanceFlags.EtcdProxy, "etcd-proxy", false, "If set, the new instance will be an ETCD proxy")
	cmdCreate.AddCommand(cmdCreateInstance)

	cmdMain.AddCommand(cmdCreate)
}

func createCluster(cmd *cobra.Command, args []string) {
	clusterInfoFromArgs(&createClusterFlags.ClusterInfo, args)

	provider := newProvider()

	// Validate
	if err := createClusterFlags.Validate(); err != nil {
		Exitf("Create failed: %s\n", err.Error())
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

	// Update all members
	isEtcdProxy := func(i providers.ClusterInstance) bool {
		return false
	}
	if err := providers.UpdateClusterMembers(log, createClusterFlags.ClusterInfo, isEtcdProxy, provider); err != nil {
		Exitf("Failed to update cluster members: %v\n", err)
	}

	Infof("Cluster created\n")
}

func createInstance(cmd *cobra.Command, args []string) {
	clusterInfoFromArgs(&createInstanceFlags.ClusterInfo, args)

	createInstanceFlags.SetupNames(createInstanceFlags.Name, createInstanceFlags.Domain)
	provider := newProvider()

	// Validate
	if err := createInstanceFlags.Validate(); err != nil {
		Exitf("Create failed: %s\n", err.Error())
	}

	// See if there are already instances for the given cluster
	instances, err := provider.GetInstances(&createInstanceFlags.ClusterInfo)
	if err != nil {
		Exitf("Failed to query existing instances: %v\n", err)
	}
	if len(instances) == 0 {
		Exitf("Cluster %s.%s does not exist.\n", createInstanceFlags.Name, createInstanceFlags.Domain)
	}

	// Create
	instance, err := provider.CreateInstance(&createInstanceFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to create new instance: %v\n", err)
	}

	// Add new instance to ETCD (if not a proxy)
	if !createInstanceFlags.EtcdProxy {
		newMachineID, err := instance.GetMachineID(log)
		if err != nil {
			Exitf("Failed to get machine ID: %v\n", err)
		}
		if err := instances[0].AddEtcdMember(log, newMachineID, instance.PrivateIpv4); err != nil {
			Exitf("Failed to add new instance to etcd: %v\n", err)
		}
	}

	// Update existing members
	isEtcdProxy := func(i providers.ClusterInstance) bool {
		return createInstanceFlags.EtcdProxy && (i.PrivateIpv4 == instance.PrivateIpv4)
	}
	if err := providers.UpdateClusterMembers(log, createInstanceFlags.ClusterInfo, isEtcdProxy, provider); err != nil {
		Exitf("Failed to update cluster members: %v\n", err)
	}

	Infof("Instance created\n")
}
