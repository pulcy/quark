package digitalocean

import (
	"github.com/pulcy/quark/providers"
)

const (
	defaultRegionID = "ams3"
	defaultImageID  = "coreos-stable"
	defaultTypeID   = "512mb"
)

// Apply defaults for the given options
func (vp *doProvider) InstanceDefaults(options providers.CreateInstanceOptions) providers.CreateInstanceOptions {
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	return options
}

// Apply defaults for the given options
func (vp *doProvider) ClusterDefaults(options providers.CreateClusterOptions) providers.CreateClusterOptions {
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
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
