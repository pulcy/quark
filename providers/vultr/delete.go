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
	"github.com/pulcy/quark/providers"
)

// Remove all instances of a cluster
func (vp *vultrProvider) DeleteCluster(info providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	servers, err := vp.getInstances(info)
	if err != nil {
		return err
	}
	for _, s := range servers {
		// Delete DNS instance records
		instance := vp.clusterInstance(s)
		if err := providers.UnRegisterInstance(vp.Logger, dnsProvider, instance, info.Domain); err != nil {
			return maskAny(err)
		}

		// Delete droplet
		err := vp.client.DeleteServer(s.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (vp *vultrProvider) DeleteInstance(info providers.ClusterInstanceInfo, dnsProvider providers.DnsProvider) error {
	fullName := info.String()
	servers, err := vp.getInstances(info.ClusterInfo)
	if err != nil {
		return err
	}
	for _, s := range servers {
		if s.Name == fullName {
			// Delete DNS instance records
			instance := vp.clusterInstance(s)
			if err := providers.UnRegisterInstance(vp.Logger, dnsProvider, instance, info.Domain); err != nil {
				return maskAny(err)
			}

			// Delete droplet
			err := vp.client.DeleteServer(s.ID)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return maskAny(NotFoundError)
}
