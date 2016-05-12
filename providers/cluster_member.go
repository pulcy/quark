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
)

type ClusterMember struct {
	ClusterID     string // ID of the cluster this is a member of (/etc/pulcu/cluster-id)
	MachineID     string // ID of the machine (/etc/machine-id)
	ClusterIP     string // IP address of the instance used for all private communication in the cluster
	PrivateHostIP string // IP address of the host on the private network (can be ClusterIP)
	EtcdProxy     bool   // If set, this member is an ETCD proxy
}

type ClusterMemberList []ClusterMember

func (cml ClusterMemberList) Render() string {
	data := ""
	for _, cm := range cml {
		options := ""
		if cm.EtcdProxy {
			options = options + " etcd-proxy"
		}
		if cm.PrivateHostIP != "" && cm.ClusterIP != cm.PrivateHostIP {
			options = options + " private-host-ip=" + cm.PrivateHostIP
		}
		data = data + fmt.Sprintf("%s=%s%s\n", cm.MachineID, cm.ClusterIP, options)
	}
	return data
}

func (cml ClusterMemberList) Find(instance ClusterInstance) (ClusterMember, error) {
	for _, cm := range cml {
		if cm.ClusterIP == instance.ClusterIP {
			return cm, nil
		}
	}
	return ClusterMember{}, maskAny(NotFoundError)
}
