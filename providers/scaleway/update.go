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
	"github.com/op/go-logging"

	"github.com/pulcy/quark/providers"
)

func (p *scalewayProvider) UpdateCluster(log *logging.Logger, info providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	instances, err := p.GetInstances(info)
	if err != nil {
		return maskAny(err)
	}
	members, err := instances.AsClusterMemberList(log, nil)
	if err != nil {
		return maskAny(err)
	}
	rebootAfter := false
	if err := instances.UpdateClusterMembers(log, members, rebootAfter, p); err != nil {
		return maskAny(err)
	}
	if err := instances.ReconfigureTincCluster(log, nil); err != nil {
		return maskAny(err)
	}
	return nil
}
