package main

import (
	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
)

var (
	cmdCreateInstance = &cobra.Command{
		Use: "create",
		Run: createInstance,
	}

	createInstanceFlags providers.CreateInstanceOptions
)

func init() {
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.Name, "name", "", "Cluster name")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.ImageID, "image", "", "OS image to run on new instances")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.RegionID, "region", "", "Region to create the instances in")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.TypeID, "type", "", "Type of the new instances")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.GluonImage, "gluon-image", defaultGluonImage, "Image containing gluon")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.RebootStrategy, "reboot-strategy", defaultRebootStrategy, "CoreOS reboot strategy")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryUrl, "private-registry-url", defaultPrivateRegistryUrl(), "URL of private docker registry")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryUserName, "private-registry-username", defaultPrivateRegistryUserName(), "Username for private registry")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryPassword, "private-registry-password", defaultPrivateRegistryPassword(), "Password for private registry")
	cmdCreateInstance.Flags().StringSliceVar(&createInstanceFlags.SSHKeyNames, "ssh-key", defaultSshKeys(), "Names of SSH keys to add to instance")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.SSHKeyGithubAccount, "ssh-key-github-account", defaultSshKeyGithubAccount(), "Github account name used to fetch SSH keys (to add to instances)")
	cmdCreateInstance.Flags().BoolVar(&createInstanceFlags.EtcdProxy, "etcd-proxy", false, "If set, the new instance will be an ETCD proxy")
	cmdInstance.AddCommand(cmdCreateInstance)
}

func createInstance(cmd *cobra.Command, args []string) {
	clusterInfoFromArgs(&createInstanceFlags.ClusterInfo, args)

	createInstanceFlags.SetupNames(createInstanceFlags.Name, createInstanceFlags.Domain)
	provider := newProvider()
	createInstanceFlags = provider.InstanceDefaults(createInstanceFlags)

	// Validate
	if err := createInstanceFlags.Validate(); err != nil {
		Exitf("Create failed: %s\n", err.Error())
	}

	// See if there are already instances for the given cluster
	instances, err := provider.GetInstances(createInstanceFlags.ClusterInfo)
	if err != nil {
		Exitf("Failed to query existing instances: %v\n", err)
	}
	if len(instances) == 0 {
		Exitf("Cluster %s.%s does not exist.\n", createInstanceFlags.Name, createInstanceFlags.Domain)
	}

	// Fetch cluster ID
	clusterID, err := instances[0].GetClusterID(log)
	if err != nil {
		Exitf("Failed to get cluster-id: %v\n", err)
	}
	createInstanceFlags.ID = clusterID

	// Fetch vault address
	vaultAddr, err := instances[0].GetVaultAddr(log)
	if err != nil {
		Exitf("Failed to get vault-addr: %v\n", err)
	}
	createInstanceFlags.VaultAddress = vaultAddr

	// Fetch vault CA certificate
	vaultCACert, err := instances[0].GetVaultCrt(log)
	if err != nil {
		Exitf("Failed to get vault-cacert: %v\n", err)
	}
	createInstanceFlags.VaultCertificate = vaultCACert

	// Create
	instance, err := provider.CreateInstance(createInstanceFlags, newDnsProvider())
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
