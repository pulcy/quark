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

package providers

import (
	"sync"

	"github.com/op/go-logging"
)

// UpdateClusterMembers updates /etc/cluster-members on all instances of the cluster
func UpdateClusterMembers(log *logging.Logger, info ClusterInfo, rebootAfter bool, isEtcdProxy func(ClusterInstance) bool, provider CloudProvider) error {
	// Load all instances
	instances, err := provider.GetInstances(info)
	if err != nil {
		return maskAny(err)
	}

	// Load cluster-members data
	clusterMembers, err := instances.AsClusterMemberList(log, isEtcdProxy)
	if err != nil {
		return maskAny(err)
	}

	// Call update-member on all instances
	if instances.UpdateClusterMembers(log, clusterMembers, rebootAfter); err != nil {
		return maskAny(err)
	}

	return nil
}

// UpdateClusterMembers updates /etc/cluster-members on all instances of the cluster
func (instances ClusterInstanceList) UpdateClusterMembers(log *logging.Logger, clusterMembers ClusterMemberList, rebootAfter bool) error {
	// Now update all members in parallel
	wg := sync.WaitGroup{}
	errorChannel := make(chan error, len(instances))
	for _, i := range instances {
		wg.Add(1)
		go func(i ClusterInstance) {
			defer wg.Done()
			if err := i.UpdateClusterMembers(log, clusterMembers); err != nil {
				errorChannel <- maskAny(err)
			}
			if rebootAfter {
				if err := i.Reboot(log); err != nil {
					errorChannel <- maskAny(err)
				}
			}
		}(i)
	}
	wg.Wait()
	close(errorChannel)
	for err := range errorChannel {
		return maskAny(err)
	}
	return nil
}
