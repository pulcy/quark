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

const (
	defaultRegionID = "ams3"
	defaultImageID  = "coreos-stable"
	defaultTypeID   = "512mb"

	privateClusterDevice = "eth1"
)

// Apply defaults for the given options
func (dp *doProvider) ClusterDefaults(options providers.ClusterInfo) providers.ClusterInfo {
	return options
}

// Apply defaults for the given options
func (dp *doProvider) CreateInstanceDefaults(options providers.CreateInstanceOptions) providers.CreateInstanceOptions {
	options.ClusterInfo = dp.ClusterDefaults(options.ClusterInfo)
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	if options.SSHKeyGithubAccount == "" {
		options.SSHKeyGithubAccount = "-"
	}
	return options
}

// Apply defaults for the given options
func (dp *doProvider) CreateClusterDefaults(options providers.CreateClusterOptions) providers.CreateClusterOptions {
	options.ClusterInfo = dp.ClusterDefaults(options.ClusterInfo)
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	if options.SSHKeyGithubAccount == "" {
		options.SSHKeyGithubAccount = "-"
	}
	return options
}

func instanceConfigDefaults(ic providers.InstanceConfig) providers.InstanceConfig {
	if ic.RegionID == "" {
		ic.RegionID = defaultRegionID
	}
	if ic.ImageID == "" {
		ic.ImageID = defaultImageID
	}
	if ic.TypeID == "" {
		ic.TypeID = defaultTypeID
	}
	return ic
}
