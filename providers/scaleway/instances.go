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
	servers, err := vp.getInstances(info)
	if err != nil {
		return nil, maskAny(err)
	}
	list := providers.ClusterInstanceList{}
	for _, s := range servers {
		info := vp.clusterInstance(s)
		list = append(list, info)

	}
	return list, nil
}

func (vp *scalewayProvider) getInstances(info providers.ClusterInfo) ([]api.ScalewayServer, error) {
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
func (dp *scalewayProvider) clusterInstance(s api.ScalewayServer) providers.ClusterInstance {
	publicIPv4 := s.PublicAddress.IP
	info := providers.ClusterInstance{
		Name:                 s.Name,
		PrivateIpv4:          s.PrivateIP,
		PublicIpv4:           publicIPv4,
		PublicIpv6:           "",
		PrivateClusterDevice: privateClusterDevice,
	}
	return info
}
