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

package cluster

import (
	"fmt"

	"github.com/juju/errgo"
)

const (
	defaultInstanceCount = 3
)

// Cluster contains all variables describing a cluster (deployment target)
type Cluster struct {
	// Name within the domain e.g. alpha-c32
	Stack string `mapstructure:"stack"`
	// Domain name e.g. pulcy.com
	Domain string `mapstructure:"domain"`
	// SSH tunnel needed to reach the cluster (optional)
	Tunnel string `mapstructure:"tunnel,omitempty"`
	// Size of the cluster (in instances==machines)
	InstanceCount int `mapstructure:"instance-count,omitempty"`
	// Default network type
	Network string `mapstructure:"network,omitempty"`

	// Quark options
	QuarkOptions QuarkOptions
}

// New returns a new cluster for testing purposes.
func New(domain, stack string, instanceCount int) Cluster {
	cluster := Cluster{
		Domain:        domain,
		Stack:         stack,
		InstanceCount: instanceCount,
	}
	cluster.setDefaults()
	return cluster
}

// validate checks the values in the given cluster
func (c Cluster) validate() error {
	if c.Stack == "" {
		return maskAny(errgo.WithCausef(nil, ValidationError, "Stack missing"))
	}
	if c.Domain == "" {
		return maskAny(errgo.WithCausef(nil, ValidationError, "Domain missing"))
	}
	if c.Tunnel == "" {
		return maskAny(errgo.WithCausef(nil, ValidationError, "Tunnel missing"))
	}
	if c.InstanceCount == 0 {
		return maskAny(errgo.WithCausef(nil, ValidationError, "InstanceCount missing"))
	} else if c.InstanceCount < 0 {
		return maskAny(errgo.WithCausef(nil, ValidationError, "InstanceCount negative"))
	}
	if err := c.QuarkOptions.validate(); err != nil {
		return maskAny(err)
	}
	return nil
}

func (c *Cluster) setDefaults() {
	if c.Tunnel == "" {
		c.Tunnel = fmt.Sprintf("%s.%s", c.Stack, c.Domain)
	}
	if c.InstanceCount == 0 {
		c.InstanceCount = defaultInstanceCount
	}
	c.QuarkOptions.setDefaults()
}

// ResolveProfile looks for a given profile. If found it merges all values into the returned value set.
// If not found it returns an error.
func (c Cluster) ResolveProfile(name string) (map[string]interface{}, error) {
	result, err := c.QuarkOptions.resolveProfile(name)
	if err != nil {
		return nil, maskAny(err)
	}
	result["domain"] = c.Domain
	result["name"] = c.Stack
	result["instance-count"] = c.InstanceCount
	return result, nil
}
