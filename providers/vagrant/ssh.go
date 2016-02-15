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

package vagrant

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

const (
	insecurePrivateKeyPathTmpl = "~/.vagrant.d/insecure_private_key"
)

func fetchVagrantInsecureSSHKey() (string, error) {
	insecurePrivateKeyPath, err := homedir.Expand(insecurePrivateKeyPathTmpl)
	if err != nil {
		return "", maskAny(err)
	}

	privateRaw, err := ioutil.ReadFile(insecurePrivateKeyPath)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", maskAny(err)
	}

	privateKey, err := ssh.ParsePrivateKey(privateRaw)
	if err != nil {
		return "", maskAny(err)
	}

	resultRaw := ssh.MarshalAuthorizedKey(privateKey.PublicKey())
	return strings.TrimSpace(string(resultRaw)), nil
}
