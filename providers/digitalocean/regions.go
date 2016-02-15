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
	"fmt"
	"sort"
	"strings"

	"github.com/ryanuber/columnize"
)

func (this *doProvider) ShowRegions() error {
	// Load regions
	client := NewDOClient(this.token)
	regions, err := RegionList(client)
	if err != nil {
		return err
	}

	lines := []string{
		"Slug | Name | Features | Size",
	}
	for _, r := range regions {
		if !r.Available {
			continue
		}
		line := fmt.Sprintf("%s | %s | %s | %s", r.Slug, r.Name, strings.Join(r.Features, " "), strings.Join(r.Sizes, " "))
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}
