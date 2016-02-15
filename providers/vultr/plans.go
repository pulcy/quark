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

	"github.com/ryanuber/columnize"
)

func (vp *vultrProvider) ShowInstanceTypes() error {
	plans, err := vp.client.GetPlans()
	if err != nil {
		return maskAny(err)
	}

	lines := []string{
		"ID | Name | VCpu | RAM | Disk | Bandwidth | Price",
	}
	for _, p := range plans {
		line := fmt.Sprintf("%03d | %s | %d | %s | %s | %s | %s", p.ID, p.Name, p.VCpus, p.RAM, p.Disk, p.Bandwidth, p.Price)
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}
