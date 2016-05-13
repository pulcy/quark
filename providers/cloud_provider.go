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
	"github.com/op/go-logging"
)

// CloudProvider holds all functions to be implemented by cloud providers
type CloudProvider interface {
	ShowRegions() error
	ShowImages() error
	ShowKeys() error
	ShowInstanceTypes() error

	// Apply defaults for the given options
	ClusterDefaults(options ClusterInfo) ClusterInfo

	// Apply defaults for the given options
	CreateInstanceDefaults(options CreateInstanceOptions) CreateInstanceOptions

	// Apply defaults for the given options
	CreateClusterDefaults(options CreateClusterOptions) CreateClusterOptions

	// Create a machine instance
	CreateInstance(log *logging.Logger, options CreateInstanceOptions, dnsProvider DnsProvider) (ClusterInstance, error)

	// Create an entire cluster
	CreateCluster(log *logging.Logger, options CreateClusterOptions, dnsProvider DnsProvider) error

	// Get names of instances of a cluster
	GetInstances(info ClusterInfo) (ClusterInstanceList, error)

	// Remove all instances of a cluster
	DeleteCluster(info ClusterInfo, dnsProvider DnsProvider) error

	// Remove a single instance of a cluster
	DeleteInstance(info ClusterInstanceInfo, dnsProvider DnsProvider) error

	// Perform a reboot of the given instance
	RebootInstance(instance ClusterInstance) error

	// Update the instances of the cluster to all new services & formats
	UpdateCluster(log *logging.Logger, info ClusterInfo, dnsProvider DnsProvider) error

	ShowDomainRecords(domain string) error
}
