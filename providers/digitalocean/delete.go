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
	"github.com/pulcy/quark/providers"
)

func (this *doProvider) DeleteCluster(info providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	droplets, err := this.getInstances(info)
	if err != nil {
		return err
	}
	client := NewDOClient(this.token)
	for _, d := range droplets {
		// Delete DNS instance records
		instance := this.clusterInstance(d)
		if err := providers.UnRegisterInstance(this.Logger, dnsProvider, instance, info.Domain); err != nil {
			return maskAny(err)
		}

		// Delete droplet
		_, err := client.Droplets.Delete(d.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dp *doProvider) DeleteInstance(info providers.ClusterInstanceInfo, dnsProvider providers.DnsProvider) error {
	fullName := info.String()
	droplets, err := dp.getInstances(info.ClusterInfo)
	if err != nil {
		return err
	}
	client := NewDOClient(dp.token)
	for _, d := range droplets {
		if d.Name == fullName {
			// Delete DNS instance records
			instance := dp.clusterInstance(d)
			if err := providers.UnRegisterInstance(dp.Logger, dnsProvider, instance, info.Domain); err != nil {
				return maskAny(err)
			}

			// Delete droplet
			_, err := client.Droplets.Delete(d.ID)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return maskAny(NotFoundError)
}
