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

package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/mitchellh/go-homedir"
)

const (
	defaultInstanceCount       = 3
	defaultGluonImage          = "pulcy/gluon:0.16.8"
	defaultRebootStrategy      = "etcd-lock"
	defaultMinOSVersion        = "835.13.0"
	defaultGithubTokenPathTmpl = "~/.pulcy/github-token"
)

func defaultDomain() string {
	return os.Getenv("QUARK_DOMAIN")
}

func defaultPrivateRegistryUrl() string {
	return os.Getenv("QUARK_REGISTRY_URL")
}

func defaultPrivateRegistryUserName() string {
	return os.Getenv("QUARK_REGISTRY_USERNAME")
}

func defaultPrivateRegistryPassword() string {
	return os.Getenv("QUARK_REGISTRY_PASSWORD")
}

func defaultSshKeys() []string {
	return []string{os.Getenv("QUARK_SSH_KEY")}
}

func defaultSshKeyGithubAccount() string {
	return os.Getenv("QUARK_SSH_KEY_GITHUB_ACCOUNT")
}

func defaultGithubToken() string {
	path, err := homedir.Expand(defaultGithubTokenPathTmpl)
	if err != nil {
		log.Warningf("Cannot expand %s: %#v", defaultGithubTokenPathTmpl, err)
		return ""
	}
	content, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return ""
	} else if err != nil {
		log.Warningf("Cannot read %s: %#v", path, err)
		return ""
	}
	return strings.TrimSpace(string(content))
}

func defaultVagrantFolder() string {
	return os.Getenv("QUARK_VAGRANT_FOLDER")
}

func defaultVaultAddr() string {
	return os.Getenv("VAULT_ADDR")
}

func defaultVaultCACert() string {
	return os.Getenv("VAULT_CACERT")
}

func defaultRegisterInstance() bool {
	v := os.Getenv("QUARK_REGISTER_INSTANCES")
	register, err := strconv.ParseBool(v)
	return (err == nil) && register
}
