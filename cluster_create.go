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
	"fmt"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
)

var (
	cmdCreateCluster = &cobra.Command{
		Use: "create",
		Run: createCluster,
	}

	createClusterFlags providers.CreateClusterOptions
)

func init() {
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.ID, "cluster-id", "", "Cluster ID (for vault-monkey)")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.Name, "name", "", "Cluster name")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.ImageID, "image", "", "OS image to run on new instances")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.RegionID, "region", "", "Region to create the instances in")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.TypeID, "type", "", "Type of the new instances")
	cmdCreateCluster.Flags().IntVar(&createClusterFlags.InstanceCount, "instance-count", defaultInstanceCount, "Number of instances in cluster")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.GluonImage, "gluon-image", defaultGluonImage, "Image containing gluon")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.RebootStrategy, "reboot-strategy", defaultRebootStrategy, "CoreOS reboot strategy")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.PrivateRegistryUrl, "private-registry-url", defaultPrivateRegistryUrl(), "URL of private docker registry")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.PrivateRegistryUserName, "private-registry-username", defaultPrivateRegistryUserName(), "Username for private registry")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.PrivateRegistryPassword, "private-registry-password", defaultPrivateRegistryPassword(), "Password for private registry")
	cmdCreateCluster.Flags().StringSliceVar(&createClusterFlags.SSHKeyNames, "ssh-key", defaultSshKeys(), "Names of SSH keys to add to instances")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.SSHKeyGithubAccount, "ssh-key-github-account", defaultSshKeyGithubAccount(), "Github account name used to fetch SSH keys (to add to instances)")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.VaultAddress, "vault-addr", defaultVaultAddr(), "URL of the vault used in this cluster")
	cmdCreateCluster.Flags().StringVar(&createClusterFlags.VaultCertificatePath, "vault-cacert", defaultVaultCACert(), "Path of the CA certificate of the vault used in this cluster")
	cmdCluster.AddCommand(cmdCreateCluster)
}

func createCluster(cmd *cobra.Command, args []string) {
	clusterInfoFromArgs(&createClusterFlags.ClusterInfo, args)

	provider := newProvider()
	createClusterFlags = provider.ClusterDefaults(createClusterFlags)

	// Create cluster ID if needed
	if createClusterFlags.ID == "" {
		createClusterFlags.ID = strings.ToLower(uniuri.NewLen(40))
	} else {
		createClusterFlags.ID = strings.ToLower(createClusterFlags.ID)
	}

	// Validate
	if err := createClusterFlags.Validate(); err != nil {
		Exitf("Create failed: %s\n", err.Error())
	}

	// See if there are already instances for the given cluster
	instances, err := provider.GetInstances(createClusterFlags.ClusterInfo)
	if err != nil {
		Exitf("Failed to query existing instances: %v\n", err)
	}
	if len(instances) > 0 {
		Exitf("Cluster %s.%s already exists.\n", createClusterFlags.Name, createClusterFlags.Domain)
	}

	// Confirm
	if err := confirm(fmt.Sprintf("Are you sure you want to create a %d instance cluster of %s?", createClusterFlags.InstanceCount, createClusterFlags.InstanceConfig)); err != nil {
		Exitf("%v\n", err)
	}

	// Create
	err = provider.CreateCluster(log, createClusterFlags, newDnsProvider())
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

	Infof("Cluster created with ID: %s\n", createClusterFlags.ID)
}
