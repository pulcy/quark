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
