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

package scaleway

import (
	"fmt"
	"strings"

	"github.com/scaleway/scaleway-cli/pkg/api"

	"github.com/pulcy/quark/providers"
)

// Get names of instances of a cluster
func (vp *scalewayProvider) GetInstances(info providers.ClusterInfo) (providers.ClusterInstanceList, error) {
	servers, err := vp.getServers(info)
	if err != nil {
		return nil, maskAny(err)
	}
	list := providers.ClusterInstanceList{}
	for _, s := range servers {
		instance := vp.clusterInstance(s, false)
		list = append(list, instance)

	}
	return list, nil
}

func (vp *scalewayProvider) getServers(info providers.ClusterInfo) ([]api.ScalewayServer, error) {
	all := true
	limit := 999
	servers, err := vp.client.GetServers(all, limit)
	if err != nil {
		return nil, maskAny(err)
	}

	postfix := fmt.Sprintf(".%s.%s", info.Name, info.Domain)
	result := []api.ScalewayServer{}
	for _, s := range *servers {
		if strings.HasSuffix(s.Name, postfix) {
			result = append(result, s)
		}
	}

	return result, nil
}

// clusterInstance creates a ClusterInstance record for the given server
func (dp *scalewayProvider) clusterInstance(s api.ScalewayServer, bootstrapNeeded bool) providers.ClusterInstance {
	publicIPv4 := s.PublicAddress.IP
	ipv6 := ""
	if s.IPV6 != nil {
		ipv6 = s.IPV6.Address
	}
	info := providers.ClusterInstance{
		ID:               s.Identifier,
		Name:             s.Name,
		ClusterIP:        s.Tags[clusterIPTagIndex],
		PrivateIP:        s.PrivateIP,
		LoadBalancerIPv4: publicIPv4,
		LoadBalancerIPv6: ipv6,
		ClusterDevice:    privateClusterDevice,
		OS:               providers.OSNameUbuntu,
	}
	if bootstrapNeeded {
		info.UserName = "root"
	}
	return info
}
