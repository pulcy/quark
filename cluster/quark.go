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

// QuarkOptions contains options that are specific to quark.
type QuarkOptions struct {
	DefaultValues map[string]interface{}
	Profiles      []Profile
}

// validate checks the values in the given cluster
func (o QuarkOptions) validate() error {
	return nil
}

func (o *QuarkOptions) setDefaults() {
}

// resolveProfile looks for a given profile. If found it merges all values into the returned value set.
// If not found it returns an error.
func (o QuarkOptions) resolveProfile(name string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range o.DefaultValues {
		result[k] = v
	}
	if name != "" {
		p, err := o.get(name)
		if err != nil {
			return nil, maskAny(err)
		}
		for k, v := range p.Values {
			result[k] = v
		}
	}
	return result, nil
}

func (o QuarkOptions) get(name string) (Profile, error) {
	for _, p := range o.Profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return Profile{}, maskAny(NotFoundError)
}
