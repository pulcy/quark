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

package vultr

import (
	"github.com/pulcy/quark/providers"
)

const (
	regionAmsterdamID = "7"
	coreosStableID    = "179"
	plan768MBID       = "29"
	plan1GBID         = "93"
)

// Apply defaults for the given options
func (vp *vultrProvider) InstanceDefaults(options providers.CreateInstanceOptions) providers.CreateInstanceOptions {
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	return options
}

// Apply defaults for the given options
func (vp *vultrProvider) ClusterDefaults(options providers.CreateClusterOptions) providers.CreateClusterOptions {
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	return options
}

func instanceConfigDefaults(ic providers.InstanceConfig) providers.InstanceConfig {
	if ic.RegionID == "" {
		ic.RegionID = regionAmsterdamID
	}
	if ic.ImageID == "" {
		ic.ImageID = coreosStableID
	}
	if ic.TypeID == "" {
		ic.TypeID = plan768MBID
	}
	return ic
}
