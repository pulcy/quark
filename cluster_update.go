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
	cmdClusterUpdate = &cobra.Command{
		Use: "update",
		Run: updateCluster,
	}

	updateClusterFlags providers.ClusterInfo
)

func init() {
	cmdClusterUpdate.Flags().StringVar(&updateClusterFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdClusterUpdate.Flags().StringVar(&updateClusterFlags.Name, "name", "", "Cluster name")
	cmdCluster.AddCommand(cmdClusterUpdate)
}

func updateCluster(cmd *cobra.Command, args []string) {
	clusterInfoFromArgs(&updateClusterFlags, args)

	provider := newProvider()
	updateClusterFlags = provider.ClusterDefaults(updateClusterFlags)

	if updateClusterFlags.Name == "" {
		Exitf("Please specify a name\n")
	}
	err := provider.UpdateCluster(log, updateClusterFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to update cluster: %v\n", err)
	}
}
