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
	"strings"

	"github.com/op/go-logging"
)

// RegisterInstance creates DNS records for an instance
func RegisterInstance(logger *logging.Logger, dnsProvider DnsProvider, options CreateInstanceOptions, name string, registerInstance, registerCluster, registerPrivateCluster bool, publicIpv4, publicIpv6, privateIpv4 string) error {
	logger.Infof("%s: '%s': '%s'", name, publicIpv4, publicIpv6)

	// Create DNS record for the instance
	logger.Infof("Creating DNS records: '%s', '%s'", options.InstanceName, options.ClusterName)
	if publicIpv4 != "" {
		if registerInstance {
			if err := dnsProvider.CreateDnsRecord(options.Domain, "A", options.InstanceName, publicIpv4); err != nil {
				return maskAny(err)
			}
		}
		if registerCluster {
			if err := dnsProvider.CreateDnsRecord(options.Domain, "A", options.ClusterName, publicIpv4); err != nil {
				return maskAny(err)
			}
		}
	}
	if privateIpv4 != "" {
		if registerPrivateCluster {
			if err := dnsProvider.CreateDnsRecord(options.Domain, "A", options.ClusterName+".private", privateIpv4); err != nil {
				return maskAny(err)
			}
		}
	}
	if publicIpv6 != "" {
		if registerInstance {
			if err := dnsProvider.CreateDnsRecord(options.Domain, "AAAA", options.InstanceName, publicIpv6); err != nil {
				return maskAny(err)
			}
		}
		if registerCluster {
			if err := dnsProvider.CreateDnsRecord(options.Domain, "AAAA", options.ClusterName, publicIpv6); err != nil {
				return maskAny(err)
			}
		}
	}

	return nil
}

// UnRegisterInstance removes DNS records for an instance
func UnRegisterInstance(logger *logging.Logger, dnsProvider DnsProvider, instance ClusterInstance, domain string) error {
	// Delete DNS instance records
	if err := dnsProvider.DeleteDnsRecord(domain, "A", instance.Name, ""); err != nil {
		return maskAny(err)
	}
	if err := dnsProvider.DeleteDnsRecord(domain, "AAAA", instance.Name, ""); err != nil {
		return maskAny(err)
	}

	// Delete DNS cluster records
	parts := strings.Split(instance.Name, ".")
	clusterName := strings.Join(parts[1:], ".")
	if instance.LoadBalancerIPv4 != "" {
		if err := dnsProvider.DeleteDnsRecord(domain, "A", clusterName, instance.LoadBalancerIPv4); err != nil {
			return maskAny(err)
		}
	}
	if instance.LoadBalancerIPv6 != "" {
		if err := dnsProvider.DeleteDnsRecord(domain, "AAAA", clusterName, instance.LoadBalancerIPv6); err != nil {
			return maskAny(err)
		}
	}

	return nil
}
