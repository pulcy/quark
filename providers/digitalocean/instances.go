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

package digitalocean

import (
	"fmt"
	"strings"

	"github.com/digitalocean/godo"

	"github.com/pulcy/quark/providers"
)

func (this *doProvider) GetInstances(info providers.ClusterInfo) ([]providers.ClusterInstance, error) {
	droplets, err := this.getInstances(info)
	if err != nil {
		return nil, err
	}
	result := []providers.ClusterInstance{}
	for _, d := range droplets {
		info := this.clusterInstance(d)
		result = append(result, info)
	}
	return result, nil
}

func (this *doProvider) getInstances(info providers.ClusterInfo) ([]godo.Droplet, error) {
	client := NewDOClient(this.token)
	droplets, err := DropletList(client)
	if err != nil {
		return nil, err
	}

	postfix := fmt.Sprintf(".%s.%s", info.Name, info.Domain)
	result := []godo.Droplet{}
	for _, d := range droplets {
		if strings.HasSuffix(d.Name, postfix) {
			result = append(result, d)
		}
	}

	return result, nil
}

// clusterInstance creates a ClusterInstance record for the given droplet
func (dp *doProvider) clusterInstance(d godo.Droplet) providers.ClusterInstance {
	info := providers.ClusterInstance{
		Name:        d.Name,
		PrivateIpv4: getIpv4(d, "private"),
		PublicIpv4:  getIpv4(d, "public"),
		PublicIpv6:  getIpv6(d, "public"),
	}
	return info
}

func getIpv4(d godo.Droplet, nType string) string {
	if d.Networks == nil {
		return ""
	}
	for _, n := range d.Networks.V4 {
		if n.Type == nType {
			return n.IPAddress
		}
	}
	return ""
}

func getIpv6(d godo.Droplet, nType string) string {
	if d.Networks == nil {
		return ""
	}
	for _, n := range d.Networks.V6 {
		if n.Type == nType {
			return n.IPAddress
		}
	}
	return ""
}
