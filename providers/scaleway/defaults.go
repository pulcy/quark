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
	"github.com/pulcy/quark/providers"
)

const (
	regionParis       = "fr-1"
	dockerImageID     = "docker"
	commercialTypeVC1 = "VC1S"
	commercialTypeC2S = "C2S"
	commercialTypeC2M = "C2M"
	commercialTypeC2L = "C2L"

	privateClusterDevice = "tun0"
	tincCIDR             = "192.168.35.0/24"
)

// Apply defaults for the given options
func (vp *scalewayProvider) ClusterDefaults(options providers.ClusterInfo) providers.ClusterInfo {
	return options
}

// Apply defaults for the given options
func (vp *scalewayProvider) CreateInstanceDefaults(options providers.CreateInstanceOptions) providers.CreateInstanceOptions {
	options.ClusterInfo = vp.ClusterDefaults(options.ClusterInfo)
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	return options
}

// Apply defaults for the given options
func (vp *scalewayProvider) CreateClusterDefaults(options providers.CreateClusterOptions) providers.CreateClusterOptions {
	options.ClusterInfo = vp.ClusterDefaults(options.ClusterInfo)
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	if options.SSHKeyGithubAccount == "" {
		options.SSHKeyGithubAccount = "-"
	}
	if options.TincCIDR == "" {
		options.TincCIDR = tincCIDR
	}
	return options
}

func instanceConfigDefaults(ic providers.InstanceConfig) providers.InstanceConfig {
	if ic.RegionID == "" {
		ic.RegionID = regionParis
	}
	if ic.ImageID == "" {
		ic.ImageID = dockerImageID
	}
	if ic.TypeID == "" {
		ic.TypeID = commercialTypeVC1
	}
	return ic
}
