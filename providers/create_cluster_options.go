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
	"net"
	"sort"
	"strings"

	"github.com/dchest/uniuri"
)

// Options for creating a cluster
type CreateClusterOptions struct {
	ClusterInfo
	InstanceConfig
	SSHKeyNames             []string // List of names of SSH keys to install on each instance
	SSHKeyGithubAccount     string   // Github account name used to fetch SSH keys
	RegisterInstance        bool     // If set, the instances will be registered with their instance name in DNS
	InstanceCount           int      // Number of instances to start
	GluonImage              string   // Docker image containing gluon
	RebootStrategy          string
	PrivateRegistryUrl      string // URL of private docker registry
	PrivateRegistryUserName string // Username of private docker registry
	PrivateRegistryPassword string // Password of private docker registry
	VaultAddress            string // URL of the vault
	VaultCertificatePath    string // Path of the vault ca-cert file
	TincCIDR                string // CIDR for the TINC network inside the cluster (e.g. 192.168.35.0/24)
	HttpProxy               string // Address of the http proxy to use (if any)

	instancePrefixes []string
}

// NewCreateInstanceOptions creates a new CreateInstanceOptions instances with all
// values inherited from the given CreateClusterOptions
func (o *CreateClusterOptions) NewCreateInstanceOptions(isCore, isLB bool, instanceIndex int) (CreateInstanceOptions, error) {
	if len(o.instancePrefixes) == 0 {
		for i := 0; i < o.InstanceCount; i++ {
			prefix := strings.ToLower(uniuri.NewLen(6))
			o.instancePrefixes = append(o.instancePrefixes, prefix)
		}
		sort.Strings(o.instancePrefixes)
	}

	/*
		raw, err := ioutil.ReadFile(o.VaultCertificatePath)
		if err != nil {
			return CreateInstanceOptions{}, maskAny(err)
		}
		vaultCertificate := string(raw)
	*/

	tincAddress := ""
	if o.TincCIDR != "" {
		tincIP, tincNet, err := net.ParseCIDR(o.TincCIDR)
		if err != nil {
			return CreateInstanceOptions{}, maskAny(err)
		}
		tincIPv4 := tincIP.To4()
		if tincIPv4 == nil {
			return CreateInstanceOptions{}, maskAny(fmt.Errorf("Expected TincCIDR to be an IPv4 CIDR, got '%s'", o.TincCIDR))
		}
		if ones, bits := tincNet.Mask.Size(); ones != 24 || bits != 32 {
			return CreateInstanceOptions{}, maskAny(fmt.Errorf("Expected TincCIDR to contain a /24 network, got '%s'", o.TincCIDR))
		}
		if instanceIndex < 1 || instanceIndex >= 255 {
			return CreateInstanceOptions{}, maskAny(fmt.Errorf("Expected instanceIndex in the range of 1..254, got %d", instanceIndex))
		}
		tincIPv4[3] = byte(instanceIndex)
		tincAddress = tincIPv4.String()
	}

	io := CreateInstanceOptions{
		ClusterInfo:             o.ClusterInfo,
		InstanceConfig:          o.InstanceConfig,
		InstanceIndex:           instanceIndex,
		RegisterInstance:        o.RegisterInstance,
		RoleCore:                isCore,
		RoleLoadBalancer:        isLB,
		SSHKeyNames:             o.SSHKeyNames,
		SSHKeyGithubAccount:     o.SSHKeyGithubAccount,
		GluonImage:              o.GluonImage,
		RebootStrategy:          o.RebootStrategy,
		PrivateRegistryUrl:      o.PrivateRegistryUrl,
		PrivateRegistryUserName: o.PrivateRegistryUserName,
		PrivateRegistryPassword: o.PrivateRegistryPassword,
		VaultAddress:            o.VaultAddress,
		VaultCertificatePath:    o.VaultCertificatePath,
		TincCIDR:                o.TincCIDR,
		TincIpv4:                tincAddress,
		HttpProxy:               o.HttpProxy,
	}
	io.SetupNames(o.instancePrefixes[instanceIndex-1], o.Name, o.Domain)
	return io, nil
}

// Validate the given options
func (cco CreateClusterOptions) Validate() error {
	if cco.Domain == "" {
		return errors.New("Please specify a domain")
	}
	if cco.Name == "" {
		return errors.New("Please specify a name")
	}
	if strings.ContainsAny(cco.Name, ".") {
		return errors.New("Invalid characters in name")
	}
	if err := cco.InstanceConfig.Validate(); err != nil {
		return maskAny(err)
	}
	if len(cco.SSHKeyNames) == 0 {
		return errors.New("Please specify at least one SSH key")
	}
	if cco.SSHKeyGithubAccount == "" {
		return errors.New("Please specify a valid ssh key github account")
	}
	if cco.InstanceCount < 1 {
		return errors.New("Please specify a valid instance count")
	}
	if cco.GluonImage == "" {
		return errors.New("Please specify a gluon-image")
	}
	if cco.PrivateRegistryUrl == "" {
		return errors.New("Please specify a private-registry-url")
	}
	if cco.PrivateRegistryUserName == "" {
		return errors.New("Please specify a private-registry-username")
	}
	if cco.PrivateRegistryPassword == "" {
		return errors.New("Please specify a private-registry-password")
	}
	if cco.VaultAddress == "" {
		return errors.New("Please specify a vault-addr")
	}
	if cco.VaultCertificatePath == "" {
		return errors.New("Please specify a vault-cacert")
	}
	return nil
}
