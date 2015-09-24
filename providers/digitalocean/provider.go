package digitalocean

import (
	"arvika.pulcy.com/iggi/droplets/providers"
)

type doProvider struct {
	token string
}

func NewProvider(token string) providers.CloudProvider {
	return &doProvider{
		token: token,
	}
}
