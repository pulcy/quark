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

type ClusterInstanceList []ClusterInstance

func (cil ClusterInstanceList) AsClusterMemberList(log *logging.Logger, isEtcdProxy func(ClusterInstance) bool) (ClusterMemberList, error) {
	wg := sync.WaitGroup{}
	errors := make(chan error, len(cil))
	memberChan := make(chan ClusterMember, len(cil))
	for _, instance := range cil {
		wg.Add(1)
		go func(instance ClusterInstance) {
			defer wg.Done()
			member, err := instance.AsClusterMember(log)
			if err != nil {
				errors <- maskAny(err)
				return
			}
			if (isEtcdProxy != nil) && isEtcdProxy(instance) {
				member.EtcdProxy = true
			}
			memberChan <- member
		}(instance)
	}
	wg.Wait()
	close(errors)
	close(memberChan)
	err := <-errors
	if err != nil {
		return nil, maskAny(err)
	}

	members := ClusterMemberList{}
	for member := range memberChan {
		members = append(members, member)
	}

	return members, nil
}

// Contains returns true if the given instance is an element of the given list, false otherwise.
func (cil ClusterInstanceList) Contains(i ClusterInstance) bool {
	for _, x := range cil {
		if x.Equals(i) {
			return true
		}
	}
	return false
}

// InstanceByName returns the instance (in the given list) with the given name.
func (cil ClusterInstanceList) InstanceByName(name string) (ClusterInstance, error) {
	for _, x := range cil {
		if x.Name == name {
			return x, nil
		}
	}
	return ClusterInstance{}, maskAny(NotFoundError)
}

// Except returns a copy of the given list except the given instance.
func (cil ClusterInstanceList) Except(i ClusterInstance) ClusterInstanceList {
	result := ClusterInstanceList{}
	for _, x := range cil {
		if !x.Equals(i) {
			result = append(result, x)
		}
	}
	return result
}
