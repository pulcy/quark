package main

import (
	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
)

const (
	defaultDomain                  = "pulcy.com"
	defaultClusterImage            = "coreos-stable"
	defaultClusterRegion           = "ams3"
	defaultClusterSize             = "512mb"
	defaultInstanceCount           = 3
	defaultYardImage               = "pulcy/yard:0.10.5"
	sshKey                         = "ewout@prangsma.net"
	defaultRebootStrategy          = "etcd-lock"
	defaultPrivateRegistryUrl      = "https://registry.pulcy.com"
	defaultPrivateRegistryUserName = "server"
	defaultPrivateRegistryPassword = ""
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
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Domain, "domain", def("QUARK_DOMAIN", "domain", defaultDomain), "Cluster domain")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Name, "name", "", "Cluster name")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Image, "image", defaultClusterImage, "OS image to run on new instances")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Region, "region", defaultClusterRegion, "Region to create the instances in")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Size, "size", defaultClusterSize, "Size of the new instances")
	cmdCreateCluster.Flags().IntVar(&createClusterFlags.InstanceCount, "instance-count", defaultInstanceCount, "Number of instances in cluster")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.YardImage, "yard-image", defaultYardImage, "Image containing encrypted yard")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.YardPassphrase, "yard-passphrase", def("", "yard-passphrase", ""), "Passphrase used to decrypt yard.gpg")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.RebootStrategy, "reboot-strategy", defaultRebootStrategy, "CoreOS reboot strategy")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.PrivateRegistryUrl, "private-registry-url", defaultPrivateRegistryUrl, "URL of private docker registry")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.PrivateRegistryUserName, "private-registry-username", def("", "private-registry-username", defaultPrivateRegistryUserName), "Username for private registry")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.PrivateRegistryPassword, "private-registry-password", def("", "private-registry-password", defaultPrivateRegistryPassword), "Password for private registry")
	cmdCreate.AddCommand(cmdCreateCluster)

	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Domain, "domain", def("QUARK_DOMAIN", "domain", defaultDomain), "Cluster domain")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Name, "name", "", "Cluster name")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Image, "image", defaultClusterImage, "OS image to run on new instances")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Region, "region", defaultClusterRegion, "Region to create the instances in")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Size, "size", defaultClusterSize, "Size of the new instances")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.YardImage, "yard-image", defaultYardImage, "Image containing encrypted yard")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.YardPassphrase, "yard-passphrase", def("", "yard-passphrase", ""), "Passphrase used to decrypt yard.gpg")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.RebootStrategy, "reboot-strategy", defaultRebootStrategy, "CoreOS reboot strategy")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryUrl, "private-registry-url", defaultPrivateRegistryUrl, "URL of private docker registry")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryUserName, "private-registry-username", def("", "private-registry-username", defaultPrivateRegistryUserName), "Username for private registry")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryPassword, "private-registry-password", def("", "private-registry-password", defaultPrivateRegistryPassword), "Password for private registry")
	cmdCreate.AddCommand(cmdCreateInstance)

	cmdMain.AddCommand(cmdCreate)
}

func createCluster(cmd *cobra.Command, args []string) {
	clusterInfoFromArgs(&createClusterFlags.ClusterInfo, args)

	createClusterFlags.SSHKeyNames = []string{sshKey}
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
	if err := providers.UpdateClusterMembers(log, createClusterFlags.ClusterInfo, provider); err != nil {
		Exitf("Failed to update cluster members: %v\n", err)
	}

	Infof("Cluster created\n")
}

func createInstance(cmd *cobra.Command, args []string) {
	clusterInfoFromArgs(&createInstanceFlags.ClusterInfo, args)

	createInstanceFlags.DiscoveryURL = "http://dummy"
	createInstanceFlags.SSHKeyNames = []string{sshKey}
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

	// Find discovery URL
	discoveryURL, err := instances[0].GetEtcdDiscoveryURL(log)
	if err != nil {
		Exitf("Failed to get discovery URL: %v\n", err)
	}
	createInstanceFlags.DiscoveryURL = discoveryURL

	// Create
	instance, err := provider.CreateInstance(&createInstanceFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to create new instance: %v\n", err)
	}

	// Add new instance to ETCD
	newMachineID, err := instance.GetMachineID(log)
	if err != nil {
		Exitf("Failed to get machine ID: %v\n", err)
	}
	if err := instances[0].AddEtcdMember(log, newMachineID, instance.PrivateIpv4); err != nil {
		Exitf("Failed to add new instance to etcd: %v\n", err)
	}

	// Update existing members
	if err := providers.UpdateClusterMembers(log, createInstanceFlags.ClusterInfo, provider); err != nil {
		Exitf("Failed to update cluster members: %v\n", err)
	}

	// Reconfigure etcd2 to connect to the existing cluster
	if err := instance.ReconfigureEtcd2(log, discoveryURL); err != nil {
		Exitf("Failed to reconfigure etcd2: %v\n", err)
	}

	Infof("Instance created\n")
}
