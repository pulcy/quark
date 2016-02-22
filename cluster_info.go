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

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
)

var (
	cmdClusterInfo = &cobra.Command{
		Use: "info",
		Run: showClusterInfo,
	}

	clusterInfoFlags providers.ClusterInfo
)

func init() {
	cmdClusterInfo.Flags().StringVar(&clusterInfoFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdClusterInfo.Flags().StringVar(&clusterInfoFlags.Name, "name", "", "Cluster name")
	cmdCluster.AddCommand(cmdClusterInfo)
}

func showClusterInfo(cmd *cobra.Command, args []string) {
	clusterInfoFromArgs(&clusterInfoFlags, args)

	provider := newProvider()
	clusterInfoFlags = provider.ClusterDefaults(clusterInfoFlags)

	if clusterInfoFlags.Name == "" {
		Exitf("Please specify a name\n")
	}
	instances, err := provider.GetInstances(instancesFlags)
	if err != nil {
		Exitf("Failed to list instances: %v\n", err)
	}
	clusterMembers, err := instances.AsClusterMemberList(log, nil)
	if err != nil {
		Exitf("Failed to fetch instance member data: %v\n", err)
	}

	lines := []string{
		fmt.Sprintf("ID | %s", clusterMembers[0].ClusterID),
		fmt.Sprintf("#Instances | %d", len(instances)),
	}
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)
}
