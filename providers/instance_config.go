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

package providers

import (
	"errors"
	"fmt"
)

type InstanceConfig struct {
	ImageID      string // ID of the image to install on each instance
	RegionID     string // ID of the region to run all instances in
	TypeID       string // ID of the type of each instance
	MinOSVersion string
	NoPublicIPv4 bool // If set, this instance will be created without a public IPv4 address
}

func (ic InstanceConfig) String() string {
	return fmt.Sprintf("type: %s, image: %s, region: %s", ic.TypeID, ic.ImageID, ic.RegionID)
}

// Validate the given options
func (ic InstanceConfig) Validate() error {
	if ic.ImageID == "" {
		return maskAny(errors.New("Please specific an image"))
	}
	if ic.RegionID == "" {
		return maskAny(errors.New("Please specific a region"))
	}
	if ic.TypeID == "" {
		return maskAny(errors.New("Please specific a type"))
	}
	return nil
}
