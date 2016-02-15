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

package vultr

import (
	"fmt"
	"sort"

	"github.com/juju/errgo"
	"github.com/ryanuber/columnize"
)

func (vp *vultrProvider) ShowKeys() error {
	keys, err := vp.client.GetSSHKeys()
	if err != nil {
		return maskAny(err)
	}
	lines := []string{
		"ID | Name | Public-key",
	}
	for _, r := range keys {
		line := fmt.Sprintf("%v | %s | %s", r.ID, r.Name, r.Key)
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}

// Search for an SSH key with given name and return its ID
func (vp *vultrProvider) findSSHKeyID(keyName string) (string, error) {
	keys, err := vp.client.GetSSHKeys()
	if err != nil {
		return "", maskAny(err)
	}
	for _, k := range keys {
		if k.Name == keyName {
			return k.ID, nil
		}
	}
	return "", errgo.WithCausef(nil, InvalidArgumentError, "key %s not found", keyName)
}
