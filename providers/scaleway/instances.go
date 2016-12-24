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
	"net"
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
	limit := 0
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
	//fmt.Printf("server=%#v\n", s)
	publicIPv4 := s.PublicAddress.IP
	ipv6 := ""
	if s.IPV6 != nil {
		ipv6 = s.IPV6.Address
	}
	privateIPMask := net.IPv4Mask(255, 255, 0, 0)
	info := providers.ClusterInstance{
		ID:               s.Identifier,
		Name:             s.Name,
		ClusterIP:        s.Tags[clusterIPTagIndex],
		PrivateIP:        s.PrivateIP,
		PrivateNetwork:   net.IPNet{IP: net.ParseIP(s.PrivateIP).Mask(privateIPMask), Mask: privateIPMask},
		PrivateDNS:       fmt.Sprintf("%s.priv.cloud.scaleway.com", s.Identifier),
		IsGateway:        publicIPv4 != "",
		LoadBalancerIPv4: publicIPv4,
		LoadBalancerIPv6: ipv6,
		LoadBalancerDNS:  fmt.Sprintf("%s.pub.cloud.scaleway.com", s.Identifier),
		ClusterDevice:    privateClusterDevice,
		OS:               providers.OSNameUbuntu,
	}
	if bootstrapNeeded {
		info.UserName = "root"
	}
	if s.PublicAddress.IP == "" {
		info.Extra = append(info.Extra, "nopubip")
	} else if *s.PublicAddress.Dynamic {
		info.Extra = append(info.Extra, "dynpubip")
		info.IsGateway = false
	}
	if s.DynamicIPRequired != nil && *s.DynamicIPRequired {
		info.Extra = append(info.Extra, "dynipreq")
	}
	return info
}
