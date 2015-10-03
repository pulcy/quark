package digitalocean

import (
	"github.com/op/go-logging"

	"arvika.pulcy.com/pulcy/droplets/providers"
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
