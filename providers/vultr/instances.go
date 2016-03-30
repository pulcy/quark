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

package vultr

import (
	"fmt"
	"strings"

	"github.com/JamesClonk/vultr/lib"

	"github.com/pulcy/quark/providers"
)

// Get names of instances of a cluster
func (vp *vultrProvider) GetInstances(info providers.ClusterInfo) (providers.ClusterInstanceList, error) {
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

func (vp *vultrProvider) getInstances(info providers.ClusterInfo) ([]lib.Server, error) {
	servers, err := vp.client.GetServers()
	if err != nil {
		return nil, maskAny(err)
	}

	postfix := fmt.Sprintf(".%s.%s", info.Name, info.Domain)
	result := []lib.Server{}
	for _, s := range servers {
		if strings.HasSuffix(s.Name, postfix) {
			result = append(result, s)
		}
	}

	return result, nil
}

// clusterInstance creates a ClusterInstance record for the given server
func (dp *vultrProvider) clusterInstance(s lib.Server) providers.ClusterInstance {
	ipv6 := ""
	if len(s.V6Networks) > 0 {
		ipv6 = s.V6Networks[0].MainIP
	}
	info := providers.ClusterInstance{
		Name:                 s.Name,
		PrivateIpv4:          s.InternalIP,
		PublicIpv4:           s.MainIP,
		PublicIpv6:           ipv6,
		PrivateClusterDevice: privateClusterDevice,
	}
	return info
}
