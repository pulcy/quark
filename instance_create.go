// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.MinOSVersion, "min-os-version", defaultMinOSVersion, "Minimum version of the OS")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.GluonImage, "gluon-image", defaultGluonImage, "Image containing gluon")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.RebootStrategy, "reboot-strategy", defaultRebootStrategy, "CoreOS reboot strategy")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryUrl, "private-registry-url", defaultPrivateRegistryUrl(), "URL of private docker registry")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryUserName, "private-registry-username", defaultPrivateRegistryUserName(), "Username for private registry")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.PrivateRegistryPassword, "private-registry-password", defaultPrivateRegistryPassword(), "Password for private registry")
	cmdCreateInstance.Flags().StringSliceVar(&createInstanceFlags.SSHKeyNames, "ssh-key", defaultSshKeys(), "Names of SSH keys to add to instance")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.SSHKeyGithubAccount, "ssh-key-github-account", defaultSshKeyGithubAccount(), "Github account name used to fetch SSH keys (to add to instances)")
	cmdCreateInstance.Flags().BoolVar(&createInstanceFlags.EtcdProxy, "etcd-proxy", false, "If set, the new instance will be an ETCD proxy")
	cmdCreateInstance.Flags().BoolVar(&createInstanceFlags.RoleCore, "role-core", false, "If set, the new instance will get `core=true` metadata")
	cmdCreateInstance.Flags().BoolVar(&createInstanceFlags.RoleLoadBalancer, "role-lb", false, "If set, the new instance will get `lb=true` metadata and register with cluster name in DNS")
	cmdCreateInstance.Flags().BoolVar(&createInstanceFlags.RoleWorker, "role-worker", false, "If set, the new instance will get `worker=true` metadata")
	cmdCreateInstance.Flags().IntVar(&createInstanceFlags.InstanceIndex, "index", 0, "Used to create `odd=true` or `even=true` metadata")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.VaultAddress, "vault-addr", defaultVaultAddr(), "URL of the vault used in this cluster")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.VaultCertificatePath, "vault-cacert", defaultVaultCACert(), "Path of the CA certificate of the vault used in this cluster")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.TincCIDR, "tinc-cidr", "", "CIDR of the TINC network in this cluster")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.TincIpv4, "tinc-ipv4", "", "IPv4 address of the new instance inside the TINC network")
	cmdCreateInstance.Flags().BoolVar(&createInstanceFlags.RegisterInstance, "register-instance", defaultRegisterInstance(), "If set, the instance will be registered with its instance name in DNS")
	cmdCreateInstance.Flags().StringVar(&createInstanceFlags.HttpProxy, "http-proxy", "", "Address of HTTP proxy to use on the instance")
	cmdInstance.AddCommand(cmdCreateInstance)
}

func createInstance(cmd *cobra.Command, args []string) {
	requireProfile := true
	loadArgumentsFromCluster(cmd.Flags(), requireProfile)
	clusterInfoFromArgs(&createInstanceFlags.ClusterInfo, args)

	provider := newProvider()
	createInstanceFlags = provider.CreateInstanceDefaults(createInstanceFlags)
	createInstanceFlags.SetupNames("", createInstanceFlags.Name, createInstanceFlags.Domain)

	// Validate
	validateVault := false
	if err := createInstanceFlags.Validate(validateVault); err != nil {
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
	createInstanceFlags.SetVaultCertificate(vaultCACert)

	// Setup instance index
	if createInstanceFlags.InstanceIndex == 0 {
		createInstanceFlags.InstanceIndex = len(instances) + 1
	}

	// Check tinc IP (if any)
	if createInstanceFlags.TincIpv4 != "" {
		for _, i := range instances {
			if i.ClusterIP == createInstanceFlags.TincIpv4 {
				Exitf("Duplicate cluster IP: %s\n", createInstanceFlags.TincIpv4)
			}
		}
	}

	// Now validate everything
	validateVault = true
	if err := createInstanceFlags.Validate(validateVault); err != nil {
		Exitf("Create failed: %s\n", err.Error())
	}

	// Create
	log.Infof("Creating new instance on %s.%s", createInstanceFlags.Name, createInstanceFlags.Domain)
	instance, err := provider.CreateInstance(log, createInstanceFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to create new instance: %v\n", err)
	}

	// Add new instance to ETCD (if not a proxy)
	if !createInstanceFlags.EtcdProxy {
		newMachineID, err := instance.GetMachineID(log)
		if err != nil {
			Exitf("Failed to get machine ID: %v\n", err)
		}
		if err := instances[0].AddEtcdMember(log, newMachineID, instance.ClusterIP); err != nil {
			Exitf("Failed to add new instance to etcd: %v\n", err)
		}
	}

	// Add new instance to list
	instances = append(instances, instance)

	// Load cluster-members data
	isEtcdProxy := func(i providers.ClusterInstance) bool {
		return createInstanceFlags.EtcdProxy && (i.ClusterIP == instance.ClusterIP)
	}
	clusterMembers, err := instances.AsClusterMemberList(log, isEtcdProxy)
	if err != nil {
		Exitf("Failed to convert instance list to member list: %v\n", err)
	}

	// Perform initial setup on new instance
	iso := providers.InitialSetupOptions{
		ClusterMembers:   clusterMembers,
		FleetMetadata:    createInstanceFlags.CreateFleetMetadata(createInstanceFlags.InstanceIndex),
		EtcdClusterState: "existing",
	}
	if err := instance.InitialSetup(log, createInstanceFlags, iso, provider); err != nil {
		Exitf("Failed to perform initial instance setup: %v\n", err)
	}

	// Update existing members
	if err := providers.UpdateClusterMembers(log, createInstanceFlags.ClusterInfo, false, isEtcdProxy, provider); err != nil {
		Exitf("Failed to update cluster members: %v\n", err)
	}

	// Reboot new instance
	if err := provider.RebootInstance(instance); err != nil {
		Exitf("Failed to reboot new instance: %v\n", err)
	}

	Infof("Instance created\n")
}
