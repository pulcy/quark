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
	"github.com/op/go-logging"
	"github.com/pulcy/vault-monkey/service"
)

type VaultProviderConfig struct {
	VaultAddr   string // URL of the vault
	VaultCACert string // Path to a PEM-encoded CA cert file to use to verify the Vault server SSL certificate
	VaultCAPath string // Path to a directory of PEM-encoded CA cert files to verify the Vault server SSL certificate
	GithubToken string
}

type VaultProvider interface {
	AddMachine(clusterId, machineId string) error
	RemoveMachine(machineId string) error
}

type vaultService struct {
	VaultProviderConfig

	log     *logging.Logger
	service *service.VaultService
}

func NewVaultProvider(log *logging.Logger, config VaultProviderConfig) (VaultProvider, error) {
	srvCfg := service.VaultServiceConfig{
		VaultAddr:   config.VaultAddr,
		VaultCACert: config.VaultCACert,
		VaultCAPath: config.VaultCAPath,
	}
	vs, err := service.NewVaultService(log, srvCfg)
	if err != nil {
		return nil, maskAny(err)
	}
	return &vaultService{
		VaultProviderConfig: config,
		log:                 log,
		service:             vs,
	}, nil
}

func (vs *vaultService) AddMachine(clusterId, machineId string) error {
	vs.log.Debug("attempting vault login")
	if err := vs.login(); err != nil {
		return maskAny(err)
	}
	c, err := vs.service.Cluster()
	if err != nil {
		return maskAny(err)
	}
	vs.log.Debugf("add machine '%s' to cluster '%s' in vault", machineId, clusterId)
	if err := c.AddMachine(clusterId, machineId, ""); err != nil {
		return maskAny(err)
	}
	return nil
}

func (vs *vaultService) RemoveMachine(machineId string) error {
	vs.log.Debug("attempting vault login")
	if err := vs.login(); err != nil {
		return maskAny(err)
	}
	c, err := vs.service.Cluster()
	if err != nil {
		return maskAny(err)
	}
	vs.log.Debugf("remove machine '%s' from vault", machineId)
	if err := c.RemoveMachine(machineId); err != nil {
		return maskAny(err)
	}
	return nil
}

func (vs *vaultService) login() error {
	loginData := service.GithubLoginData{
		GithubToken: vs.GithubToken,
	}
	if err := vs.service.GithubLogin(loginData); err != nil {
		return maskAny(err)
	}
	return nil
}
