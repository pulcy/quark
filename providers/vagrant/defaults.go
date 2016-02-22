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

package vagrant

import (
	"github.com/pulcy/quark/providers"
)

const (
	privateClusterDevice = "eth1"
)

// Apply defaults for the given options
func (vp *vagrantProvider) InstanceDefaults(options providers.CreateInstanceOptions) providers.CreateInstanceOptions {
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	return options
}

// Apply defaults for the given options
func (vp *vagrantProvider) ClusterDefaults(options providers.CreateClusterOptions) providers.CreateClusterOptions {
	if options.Name == "" {
		options.Name = "vagrant"
	}
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	return options
}

func instanceConfigDefaults(ic providers.InstanceConfig) providers.InstanceConfig {
	if ic.RegionID == "" {
		ic.RegionID = "local"
	}
	if ic.ImageID == "" {
		ic.ImageID = "coreos-stable"
	}
	if ic.TypeID == "" {
		ic.TypeID = "n/a"
	}
	return ic
}
