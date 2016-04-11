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
	"github.com/scaleway/scaleway-cli/pkg/api"

	"github.com/pulcy/quark/providers"
)

// Remove all instances of a cluster
func (vp *scalewayProvider) DeleteCluster(info providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	servers, err := vp.getServers(info)
	if err != nil {
		return err
	}
	for _, s := range servers {
		if err := vp.deleteServer(s, dnsProvider, info.Domain); err != nil {
			return maskAny(err)
		}
	}

	return nil
}

func (vp *scalewayProvider) DeleteInstance(info providers.ClusterInstanceInfo, dnsProvider providers.DnsProvider) error {
	fullName := info.String()
	servers, err := vp.getServers(info.ClusterInfo)
	if err != nil {
		return maskAny(err)
	}

	found := false
	for _, s := range servers {
		if s.Name == fullName {
			if err := vp.deleteServer(s, dnsProvider, info.Domain); err != nil {
				return maskAny(err)
			}
			found = true
			break
		}
	}

	if !found {
		return maskAny(NotFoundError)
	}

	// Reconfigure tinc
	instanceList, err := vp.GetInstances(info.ClusterInfo)
	if err != nil {
		return maskAny(err)
	}
	// Create tinc network config
	newInstances := providers.ClusterInstanceList{}
	if instanceList.ReconfigureTincCluster(vp.Logger, newInstances); err != nil {
		return maskAny(err)
	}

	return nil
}

func (vp *scalewayProvider) deleteServer(s api.ScalewayServer, dnsProvider providers.DnsProvider, domain string) error {
	if s.State == "running" {
		vp.Logger.Infof("Stopping server %s", s.Name)
		if err := vp.client.PostServerAction(s.Identifier, "terminate"); err != nil {
			return maskAny(err)
		}
		api.WaitForServerStopped(vp.client, s.Identifier)
	} else {
		vp.Logger.Infof("Server %s is at state '%s'", s.Name, s.State)
	}

	// Delete DNS instance records
	vp.Logger.Infof("Unregistering DNS for %s", s.Name)
	instance := vp.clusterInstance(s, false)
	if err := providers.UnRegisterInstance(vp.Logger, dnsProvider, instance, domain); err != nil {
		return maskAny(err)
	}

	// Delete server
	/*err := vp.client.DeleteServer(s.Identifier)
	if err != nil {
		vp.Logger.Errorf("Failed to delete server %s: %#v", s.Name, err)
		return maskAny(err)
	}*/

	// Delete volume
	/*for _, v := range s.Volumes {
		if err := vp.client.DeleteVolume(v.Identifier); err != nil {
		vp.Logger.Errorf("Failed to delete volume %s: %#v", v.Identifier, err)
			return maskAny(err)
		}
	}*/

	// Delete IP
	if !(*s.PublicAddress.Dynamic) {
		if err := vp.client.DeleteIP(s.PublicAddress.Identifier); err != nil {
			vp.Logger.Errorf("Failed to delete IP %s: %#v", s.PublicAddress.Identifier, err)
			return maskAny(err)
		}
	}

	return nil
}
