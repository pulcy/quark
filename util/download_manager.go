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

package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

type DownloadManager struct {
}

// Download downloads the file from the given URL and stores it in a local file.
// It returns the path to the local file.
func (dm *DownloadManager) Download(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", maskAny(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", maskAny(fmt.Errorf("invalid status: %d", resp.StatusCode))
	}
	tmp, err := ioutil.TempFile("", "quark")
	if err != nil {
		return "", maskAny(err)
	}
	path := tmp.Name()
	defer resp.Body.Close()
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return "", maskAny(err)
	}
	tmp.Close()
	return path, nil
}
