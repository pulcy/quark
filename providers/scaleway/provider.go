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
	"github.com/op/go-logging"
	"github.com/scaleway/scaleway-cli/pkg/api"

	"github.com/pulcy/quark/providers"
)

type scalewayProvider struct {
	Logger       *logging.Logger
	client       *api.ScalewayAPI
	organization string
}

// NewProvider creates a new Scaleway provider implementation
func NewProvider(logger *logging.Logger, organization, token string) (providers.CloudProvider, error) {
	client, err := api.NewScalewayAPI(organization, token, "quark")
	if err != nil {
		return nil, maskAny(err)
	}
	return &scalewayProvider{
		Logger:       logger,
		client:       client,
		organization: organization,
	}, nil
}
