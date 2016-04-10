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
	"fmt"

	"github.com/op/go-logging"
	"github.com/scaleway/scaleway-cli/pkg/api"

	"github.com/pulcy/quark/providers"
)

// ScalewayProviderConfig contains scaleway specific provider configuration
type ScalewayProviderConfig struct {
	// Authentication
	Organization string
	Token        string

	ReserveLoadBalancerIP bool // If set, a reserved IP address will be used for the public IPv4 address
	EnableIPV6            bool // If set, an IPv6 address will be used
}

type scalewayProvider struct {
	ScalewayProviderConfig
	Logger *logging.Logger
	client *api.ScalewayAPI
}

// NewConfig initializes a default set of provider configuration options
func NewConfig() ScalewayProviderConfig {
	return ScalewayProviderConfig{
		ReserveLoadBalancerIP: true,
		EnableIPV6:            true,
	}
}

// NewProvider creates a new Scaleway provider implementation
func NewProvider(logger *logging.Logger, config ScalewayProviderConfig) (providers.CloudProvider, error) {
	if config.Organization == "" {
		return nil, maskAny(fmt.Errorf("Organization not set"))
	}
	if config.Token == "" {
		return nil, maskAny(fmt.Errorf("Token not set"))
	}
	client, err := api.NewScalewayAPI(config.Organization, config.Token, "quark")
	if err != nil {
		return nil, maskAny(err)
	}
	return &scalewayProvider{
		ScalewayProviderConfig: config,
		Logger:                 logger,
		client:                 client,
	}, nil
}
