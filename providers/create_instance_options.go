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
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/dchest/uniuri"
)

// CreateInstanceOptions contains all options for creating an instance
type CreateInstanceOptions struct {
	ClusterInfo
	InstanceConfig
	ClusterName             string   // Full name of the cluster e.g. "dev1.example.com"
	InstanceName            string   // Name of the instance e.g. "abc123.dev1.example.com"
	InstanceIndex           int      // 0,... used for odd/even metadata
	RegisterInstance        bool     // If set, the instance will be register with its instance name in DNS
	RoleCore                bool     // If set, this instance will get `core=true` metadata
	RoleLoadBalancer        bool     // If set, this instance will get `lb=true` metadata and the instance will be registered under the cluster name in DNS
	RoleVault               bool     // If set, this instance will get `vault=true` metadata and a `vault` role.
	RoleWorker              bool     // If set, this instance will get `worker=true` metadata
	SSHKeyNames             []string // List of names of SSH keys to install
	SSHKeyGithubAccount     string   // Github account name used to fetch SSH keys
	GluonImage              string   // Docker image containing gluon
	GluonEnv                string   // Content of gluon.env
	RebootStrategy          string
	PrivateRegistryUrl      string // URL of private docker registry
	PrivateRegistryUserName string // Username of private docker registry
	PrivateRegistryPassword string // Password of private docker registry
	EtcdProxy               bool   // If set, this instance will be an ETCD proxy
	VaultAddress            string // URL of the vault
	VaultCertificatePath    string // Path of the vault ca-cert file
	vaultCertificate        string // Contents of the vault ca-cert
	VaultServerKeyPath      string // Path of the vault ca-cert key file
	VaultServerKeyCommand   string // Shell command that outputs a PEM-encoded CA key to use to as the Vault server SSL certificate key
	vaultServerKey          string // Contents of the vault ca-cert key file
	TincCIDR                string // CIDR for the TINC network inside the cluster (e.g. 192.168.35.0/24)
	TincIpv4                string // IP addres of tun0 (tinc) on this instance
	HttpProxy               string // Address of the http proxy to use (if any)
	WeaveEnv                string // Content of weave.env
	WeaveSeed               string // Content of weave-seed
}

// SetupNames configured the ClusterName and InstanceName of the given options
// using the given cluster & domain name
func (o *CreateInstanceOptions) SetupNames(prefix, clusterName, domain string) {
	if prefix == "" {
		prefix = strings.ToLower(uniuri.NewLen(6))
	}
	o.ClusterName = fmt.Sprintf("%s.%s", clusterName, domain)
	o.InstanceName = fmt.Sprintf("%s.%s.%s", prefix, clusterName, domain)
}

// VaultCertificate reads the VaultCertificatePath and returns its content as a string
func (o *CreateInstanceOptions) VaultCertificate() (string, error) {
	if o.vaultCertificate == "" {
		raw, err := ioutil.ReadFile(o.VaultCertificatePath)
		if err != nil {
			return "", maskAny(err)
		}
		o.vaultCertificate = string(raw)
	}
	return o.vaultCertificate, nil
}

// VaultServerKey reads the VaultServerKeyPath or executes the VaultServerKeyCommand and returns its content as a string
func (o *CreateInstanceOptions) VaultServerKey() (string, error) {
	if o.vaultServerKey == "" {
		if o.VaultServerKeyPath != "" {
			raw, err := ioutil.ReadFile(o.VaultServerKeyPath)
			if err != nil {
				return "", maskAny(err)
			}
			o.vaultServerKey = string(raw)
		} else if o.VaultServerKeyCommand != "" {
			cmd := exec.Command("sh", "-c", o.VaultServerKeyCommand)
			output, err := cmd.Output()
			if err != nil {
				return "", maskAny(err)
			}
			o.vaultServerKey = string(output)
		}
	}
	return o.vaultServerKey, nil
}

// SetVaultCertificate sets the content of the VaultCertificate
func (o *CreateInstanceOptions) SetVaultCertificate(contents string) {
	o.vaultCertificate = contents
}

// NewCloudConfigOptions creates a new CloudConfigOptions instances with all
// values inherited from the given CreateInstanceOptions
func (o *CreateInstanceOptions) NewCloudConfigOptions() CloudConfigOptions {
	cco := CloudConfigOptions{
		ClusterID:      o.ClusterInfo.ID,
		RebootStrategy: o.RebootStrategy,
	}
	return cco
}

// CreateFleetMetadata creates a valid fleet metadata string for use in cloud-config
func (o *CreateInstanceOptions) CreateFleetMetadata(instanceIndex int) string {
	list := []string{fmt.Sprintf("region=%s", o.RegionID)}
	if instanceIndex%2 == 0 {
		list = append(list, "even=true")
	} else {
		list = append(list, "odd=true")
	}
	if o.RoleCore {
		list = append(list, "core=true")
	}
	if o.RoleLoadBalancer {
		list = append(list, "lb=true")
	}
	if o.RoleVault {
		list = append(list, "vault=true")
	}
	if o.RoleWorker {
		list = append(list, "worker=true")
	}
	return strings.Join(list, ",")
}

// Roles returns the roles that the instance is supposed to play.
func (o *CreateInstanceOptions) Roles() string {
	var list []string
	if o.RoleCore {
		list = append(list, "core")
	}
	if o.RoleLoadBalancer {
		list = append(list, "lb")
	}
	if o.RoleVault {
		list = append(list, "vault")
	}
	if o.RoleWorker {
		list = append(list, "worker")
	}
	return strings.Join(list, ",")
}

// Validate the given options
func (cio CreateInstanceOptions) Validate(validateVault, validateWeave bool) error {
	if cio.ClusterName == "" {
		return errors.New("Please specify a cluster-name")
	}
	if cio.InstanceName == "" {
		return errors.New("Please specify a instance-name")
	}
	if err := cio.InstanceConfig.Validate(); err != nil {
		return maskAny(err)
	}
	if len(cio.SSHKeyNames) == 0 {
		return errors.New("Please specific at least one SSH key")
	}
	if cio.SSHKeyGithubAccount == "" {
		return errors.New("Please specify a valid ssh key github account")
	}
	if cio.GluonImage == "" {
		return errors.New("Please specify a gluon-image")
	}
	if validateVault {
		if cio.VaultAddress == "" {
			return errors.New("Please specify a vault-addr")
		}
		if content, err := cio.VaultCertificate(); err != nil {
			return maskAny(err)
		} else if content == "" {
			return errors.New("Please specify a vault-cacert")
		}
		if cio.RoleVault {
			if content, err := cio.VaultServerKey(); err != nil {
				return maskAny(err)
			} else if content == "" {
				return errors.New("Please specify a vault-key")
			}
		}
	}
	if validateWeave {
		if cio.WeaveEnv == "" {
			return errors.New("Please specify a weave.env")
		}
		// Note WeaveSeed is allowed to be empty
	}
	return nil
}
