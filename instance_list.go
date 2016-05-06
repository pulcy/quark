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

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
)

var (
	cmdInstanceList = &cobra.Command{
		Use: "list",
		Run: showInstances,
	}

	instancesFlags providers.ClusterInfo
)

func init() {
	cmdInstanceList.Flags().StringVar(&instancesFlags.Domain, "domain", defaultDomain(), "Cluster domain")
	cmdInstanceList.Flags().StringVar(&instancesFlags.Name, "name", "", "Cluster name")
	cmdInstance.AddCommand(cmdInstanceList)
}

func showInstances(cmd *cobra.Command, args []string) {
	loadArgumentsFromCluster(cmd.Flags())
	clusterInfoFromArgs(&instancesFlags, args)

	provider := newProvider()
	instancesFlags = provider.ClusterDefaults(instancesFlags)

	if instancesFlags.Name == "" {
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

	lines := []string{"Name | Cluster IP | Public IP | Private IP | Machine ID | Options | Extra"}
	for _, i := range instances {
		cm, _ := clusterMembers.Find(i) // ignore errors
		options := []string{}
		if cm.EtcdProxy {
			options = append(options, "etcd-proxy")
		}
		lbIP := strings.TrimSpace(i.LoadBalancerIPv4 + " " + i.LoadBalancerIPv6)
		lines = append(lines, fmt.Sprintf("%s | %s | %s | %s | %s | %s | %s", i.Name, i.ClusterIP, lbIP, i.PrivateIP, cm.MachineID, strings.Join(options, ","), strings.Join(i.Extra, ",")))
	}
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)
}
