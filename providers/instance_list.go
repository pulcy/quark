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
	"fmt"
	"net"
	"sync"

	"github.com/op/go-logging"
)

type ClusterInstanceList []ClusterInstance

func (cil ClusterInstanceList) AsClusterMemberList(log *logging.Logger, isEtcdProxy func(ClusterInstance) (bool, error)) (ClusterMemberList, error) {
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
			if isEtcdProxy != nil {
				etcdProxy, err := isEtcdProxy(instance)
				if err != nil {
					errors <- maskAny(err)
					return
				}
				member.EtcdProxy = etcdProxy
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

// IsFreeClusterIP returns true if the given IP address is not used as a cluster IP by any of the instances.
// false otherwise.
func (cil ClusterInstanceList) IsFreeClusterIP(ip net.IP) bool {
	ipAddr := ip.String()
	for _, x := range cil {
		if x.ClusterIP == ipAddr {
			return false
		}
	}
	return true
}

// CreateClusterIP returns an IP address in the given CIDR, not used by any of the instances.
func (cil ClusterInstanceList) CreateClusterIP(cidr string) (net.IP, error) {
	ip, nw, err := net.ParseCIDR(cidr)
	if err != nil {
		return net.IP{}, maskAny(err)
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return net.IP{}, maskAny(fmt.Errorf("Expected CIDR to be an IPv4 CIDR, got '%s'", cidr))
	}
	if ones, bits := nw.Mask.Size(); ones != 24 || bits != 32 {
		return net.IP{}, maskAny(fmt.Errorf("Expected CIDR to contain a /24 network, got '%s'", cidr))
	}
	lastFreeIndex := -1
	for i := 254; i >= 1; i-- {
		ipv4[3] = byte(i)
		if cil.IsFreeClusterIP(ipv4) {
			lastFreeIndex = i
		} else if lastFreeIndex > 0 {
			ipv4[3] = byte(lastFreeIndex)
			return ipv4, nil
		}
	}
	for i := 1; i < 255; i++ {
		ipv4[3] = byte(i)
		if cil.IsFreeClusterIP(ipv4) {
			return ipv4, nil
		}
	}
	return net.IP{}, maskAny(fmt.Errorf("no free IP left in '%s'", cidr))
}

// GetClusterID loads the cluster ID from any of the instances in the given list
func (cil ClusterInstanceList) GetClusterID(log *logging.Logger) (string, error) {
	for _, i := range cil {
		s, err := i.Connect()
		if err != nil {
			return "", maskAny(err)
		}
		defer s.Close()
		result, err := s.GetClusterID(log)
		if err == nil {
			return result, nil
		}
		log.Warningf("cannot get cluster IP from '%s': %#v", i, err)
	}
	return "", maskAny(fmt.Errorf("cannot get cluster IP"))
}

// GetVaultCrt loads the vault certificate from any of the instances in the given list
func (cil ClusterInstanceList) GetVaultCrt(log *logging.Logger) (string, error) {
	for _, i := range cil {
		s, err := i.Connect()
		if err != nil {
			return "", maskAny(err)
		}
		defer s.Close()
		result, err := s.GetVaultCrt(log)
		if err == nil {
			return result, nil
		}
		log.Warningf("cannot get vault certificate from '%s': %#v", i, err)
	}
	return "", maskAny(fmt.Errorf("cannot get vault certificate"))
}

// GetVaultAddr loads the vault address from any of the instances in the given list
func (cil ClusterInstanceList) GetVaultAddr(log *logging.Logger) (string, error) {
	for _, i := range cil {
		s, err := i.Connect()
		if err != nil {
			return "", maskAny(err)
		}
		defer s.Close()
		result, err := s.GetVaultAddr(log)
		if err == nil {
			return result, nil
		}
		log.Warningf("cannot get vault address from '%s': %#v", i, err)
	}
	return "", maskAny(fmt.Errorf("cannot get vault address"))
}

func (cil ClusterInstanceList) GetGluonEnv(log *logging.Logger) (string, error) {
	for _, i := range cil {
		s, err := i.Connect()
		if err != nil {
			return "", maskAny(err)
		}
		defer s.Close()
		result, err := s.GetGluonEnv(log)
		if err == nil {
			return result, nil
		}
		log.Warningf("cannot get gluon.env from '%s': %#v", i, err)
	}
	return "", maskAny(fmt.Errorf("cannot get gluon.env"))
}

func (cil ClusterInstanceList) GetWeaveEnv(log *logging.Logger) (string, error) {
	for _, i := range cil {
		s, err := i.Connect()
		if err != nil {
			return "", maskAny(err)
		}
		defer s.Close()
		result, err := s.GetWeaveEnv(log)
		if err == nil {
			return result, nil
		}
		log.Warningf("cannot get weave.env from '%s': %#v", i, err)
	}
	return "", maskAny(fmt.Errorf("cannot get weave.env"))
}

func (cil ClusterInstanceList) GetWeaveSeed(log *logging.Logger) (string, error) {
	for _, i := range cil {
		s, err := i.Connect()
		if err != nil {
			return "", maskAny(err)
		}
		defer s.Close()
		result, err := s.GetWeaveSeed(log)
		if err == nil {
			return result, nil
		}
		log.Warningf("cannot get weave-seed from '%s': %#v", i, err)
	}
	return "", maskAny(fmt.Errorf("cannot get weave-seed"))
}

// AddEtcdMember calls etcdctl to add a member to ETCD on any of the instances in the given list
func (cil ClusterInstanceList) AddEtcdMember(log *logging.Logger, name, clusterIP string) error {
	for _, i := range cil {
		s, err := i.Connect()
		if err != nil {
			return maskAny(err)
		}
		defer s.Close()
		if err := s.AddEtcdMember(log, name, clusterIP); err == nil {
			return nil
		}
		log.Warningf("cannot add '%s' to ETCD: %#v", name, err)
	}
	return maskAny(fmt.Errorf("cannot add '%s' to ETCD", name))
}

// RemoveEtcdMember calls etcdctl to remove a member from ETCD on any of the instances in the given list
func (cil ClusterInstanceList) RemoveEtcdMember(log *logging.Logger, name, clusterIP string) error {
	for _, i := range cil {
		s, err := i.Connect()
		if err != nil {
			return maskAny(err)
		}
		defer s.Close()
		if err := s.RemoveEtcdMember(log, name, clusterIP); err == nil {
			return nil
		}
		log.Warningf("cannot remove '%s' from ETCD: %#v", name, err)
	}
	return maskAny(fmt.Errorf("cannot remove '%s' from ETCD", name))
}
