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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ScalewayRC struct {
	Organization string `json:"organization"`
	Token        string `json:"token"`
}

func ReadRC() (ScalewayRC, error) {
	home := os.Getenv("HOME")
	data, err := ioutil.ReadFile(filepath.Join(home, ".scwrc"))
	if err != nil {
		return ScalewayRC{}, maskAny(err)
	}
	var rc ScalewayRC
	if err := json.Unmarshal(data, &rc); err != nil {
		return ScalewayRC{}, maskAny(err)
	}
	return rc, nil
}
