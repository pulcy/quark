package vultr

import (
	"github.com/JamesClonk/vultr/lib"
	"github.com/op/go-logging"

	"github.com/pulcy/quark/providers"
)

type vultrProvider struct {
	Logger *logging.Logger
	client *lib.Client
}

// NewProvider creates a new Vultr provider implementation
func NewProvider(logger *logging.Logger, apiKey string) providers.CloudProvider {
	client := lib.NewClient(apiKey, nil)
	return &vultrProvider{
		Logger: logger,
		client: client,
	}
}
