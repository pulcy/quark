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
