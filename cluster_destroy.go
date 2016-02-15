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

	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
)

var (
	cmdDestroyCluster = &cobra.Command{
		Short: "Destroy an entire cluster",
		Long:  "Destroy an entire cluster",
		Use:   "destroy",
		Run:   destroyCluster,
	}

	destroyClusterFlags providers.ClusterInfo
)

func init() {
	cmdDestroyCluster.Flags().StringVar(&destroyClusterFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdDestroyCluster.Flags().StringVar(&destroyClusterFlags.Name, "name", "", "Cluster name")
	cmdCluster.AddCommand(cmdDestroyCluster)
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
	err := provider.DeleteCluster(destroyClusterFlags, newDnsProvider())
	if err != nil {
		Exitf("Failed to destroy cluster: %v\n", err)
	}
}
