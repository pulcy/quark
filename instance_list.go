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
	clusterInfoFromArgs(&instancesFlags, args)

	if instancesFlags.Name == "" {
		Exitf("Please specify a name\n")
	}
	provider := newProvider()
	instances, err := provider.GetInstances(instancesFlags)
	if err != nil {
		Exitf("Failed to list instances: %v\n", err)
	}
	lines := []string{"Name | Private IP | Public IP"}
	for _, i := range instances {
		lines = append(lines, fmt.Sprintf("%s | %s | %s", i.Name, i.PrivateIpv4, i.PublicIpv4))
	}
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)
}
