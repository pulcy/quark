package vagrant

import (
	"github.com/pulcy/quark/providers"
)

// Apply defaults for the given options
func (vp *vagrantProvider) InstanceDefaults(options providers.CreateInstanceOptions) providers.CreateInstanceOptions {
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	return options
}

// Apply defaults for the given options
func (vp *vagrantProvider) ClusterDefaults(options providers.CreateClusterOptions) providers.CreateClusterOptions {
	options.InstanceConfig = instanceConfigDefaults(options.InstanceConfig)
	return options
}

func instanceConfigDefaults(ic providers.InstanceConfig) providers.InstanceConfig {
	if ic.RegionID == "" {
		ic.RegionID = "n/a"
	}
	if ic.ImageID == "" {
		ic.ImageID = "n/a"
	}
	if ic.TypeID == "" {
		ic.TypeID = "n/a"
	}
	return ic
}
