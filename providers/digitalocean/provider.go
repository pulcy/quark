package digitalocean

import (
	"github.com/op/go-logging"

	"github.com/pulcy/quark/providers"
)

type doProvider struct {
	Logger *logging.Logger
	token  string
}

func NewProvider(logger *logging.Logger, token string) providers.CloudProvider {
	return &doProvider{
		Logger: logger,
		token:  token,
	}
}

func (vp *doProvider) ShowInstanceTypes() error {
	return maskAny(NotImplementedError)
}

// Apply defaults for the given options
func (vp *doProvider) InstanceDefaults(options providers.CreateInstanceOptions) providers.CreateInstanceOptions {
	return options
}

// Apply defaults for the given options
func (vp *doProvider) ClusterDefaults(options providers.CreateClusterOptions) providers.CreateClusterOptions {
	return options
}
