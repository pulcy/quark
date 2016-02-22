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
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/juju/errgo"
	"github.com/op/go-logging"
)

const (
	defaultUserName = "core"
)

func (i ClusterInstance) runRemoteCommand(log *logging.Logger, command, stdin string, quiet bool) (string, error) {
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", defaultUserName+"@"+i.PublicIpv4, command)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	if err := cmd.Run(); err != nil {
		if !quiet {
			log.Errorf("SSH failed: %s %s", cmd.Path, strings.Join(cmd.Args, " "))
		}
		return "", errgo.NoteMask(err, stdErr.String())
	}

	out := stdOut.String()
	out = strings.TrimSuffix(out, "\n")
	return out, nil
}

func (i ClusterInstance) GetClusterID(log *logging.Logger) (string, error) {
	log.Debug("Fetching cluster-id on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "sudo cat /etc/pulcy/cluster-id", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetMachineID(log *logging.Logger) (string, error) {
	log.Debug("Fetching machine-id on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "cat /etc/machine-id", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetVaultCrt(log *logging.Logger) (string, error) {
	log.Debug("Fetching vault.crt on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "sudo cat /etc/pulcy/vault.crt", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetVaultAddr(log *logging.Logger) (string, error) {
	const prefix = "VAULT_ADDR="
	log.Debug("Fetching vault-addr on %s", i.PublicIpv4)
	env, err := i.runRemoteCommand(log, "sudo cat /etc/pulcy/vault.env", "", false)
	if err != nil {
		return "", maskAny(err)
	}
	for _, line := range strings.Split(env, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(line[len(prefix):]), nil
		}
	}
	return "", maskAny(errgo.New("VAULT_ADDR not found in /etc/pulcy/vault.env"))
}

func (i ClusterInstance) IsEtcdProxy(log *logging.Logger) (bool, error) {
	log.Debug("Fetching etcd proxy status on %s", i.PublicIpv4)
	cat, err := i.runRemoteCommand(log, "systemctl cat etcd2.service", "", false)
	return strings.Contains(cat, "ETCD_PROXY"), maskAny(err)
}

// AddEtcdMember calls etcdctl to add a member to ETCD
func (i ClusterInstance) AddEtcdMember(log *logging.Logger, name, privateIP string) error {
	log.Info("Adding %s(%s) to etcd on %s", name, privateIP, i.PublicIpv4)
	cmd := []string{
		"etcdctl",
		"member",
		"add",
		name,
		fmt.Sprintf("http://%s:2380", privateIP),
	}
	if _, err := i.runRemoteCommand(log, strings.Join(cmd, " "), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// RemoveEtcdMember calls etcdctl to remove a member from ETCD
func (i ClusterInstance) RemoveEtcdMember(log *logging.Logger, name, privateIP string) error {
	log.Info("Removing %s(%s) from etcd on %s", name, privateIP, i.PublicIpv4)
	id, err := i.runRemoteCommand(log, fmt.Sprintf("sh -c 'etcdctl member list | grep %s | cut -d: -f1'", privateIP), "", false)
	if err != nil {
		return maskAny(err)
	}
	cmd := []string{
		"etcdctl",
		"member",
		"remove",
		id,
	}
	if _, err := i.runRemoteCommand(log, strings.Join(cmd, " "), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

func (i ClusterInstance) AsClusterMember(log *logging.Logger) (ClusterMember, error) {
	clusterID, err := i.GetClusterID(log)
	if err != nil {
		return ClusterMember{}, maskAny(err)
	}
	machineID, err := i.GetMachineID(log)
	if err != nil {
		return ClusterMember{}, maskAny(err)
	}
	etcdProxy, err := i.IsEtcdProxy(log)
	if err != nil {
		return ClusterMember{}, maskAny(err)
	}
	return ClusterMember{
		ClusterID: clusterID,
		MachineID: machineID,
		PrivateIP: i.PrivateIpv4,
		EtcdProxy: etcdProxy,
	}, nil
}

type InitialSetupOptions struct {
	ClusterMembers ClusterMemberList
	FleetMetadata  string
}

// InitialSetup creates initial files and calls gluon for the first time
func (i ClusterInstance) InitialSetup(log *logging.Logger, cio CreateInstanceOptions, iso InitialSetupOptions) error {
	if _, err := i.runRemoteCommand(log, "sudo /usr/bin/mkdir -p /etc/pulcy", "", false); err != nil {
		return maskAny(err)
	}
	data := iso.ClusterMembers.Render()
	if _, err := i.runRemoteCommand(log, "sudo tee /etc/pulcy/cluster-members", data, false); err != nil {
		return maskAny(err)
	}

	vaultEnv := []string{
		fmt.Sprintf("VAULT_ADDR=%s", cio.VaultAddress),
		fmt.Sprintf("VAULT_CACERT=/etc/pulcy/vault.crt"),
	}
	if _, err := i.runRemoteCommand(log, "sudo tee /etc/pulcy/vault.env", strings.Join(vaultEnv, "\n"), false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo chmod 0400 /etc/pulcy/vault.env", "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo tee /etc/pulcy/vault.crt", cio.VaultCertificate, false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo chmod 0400 /etc/pulcy/vault.crt", "", false); err != nil {
		return maskAny(err)
	}

	log.Info("Downloading gluon on %s", i.PublicIpv4)
	if _, err := i.runRemoteCommand(log, "sudo /usr/bin/mkdir -p /home/core/bin", "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("docker run --rm -v /home/core/bin:/destination/ %s", cio.GluonImage), "", false); err != nil {
		return maskAny(err)
	}
	log.Info("Running gluon on %s", i.PublicIpv4)
	gluonArgs := []string{
		fmt.Sprintf("--gluon-image=%s", cio.GluonImage),
		fmt.Sprintf("--docker-ip=%s", i.PrivateIpv4),
		fmt.Sprintf("--private-ip=%s", i.PrivateIpv4),
		fmt.Sprintf("--private-cluster-device=%s", i.PrivateClusterDevice),
		fmt.Sprintf("--private-registry-url=%s", cio.PrivateRegistryUrl),
		fmt.Sprintf("--private-registry-username=%s", cio.PrivateRegistryUserName),
		fmt.Sprintf("--private-registry-password=%s", cio.PrivateRegistryPassword),
		fmt.Sprintf("--fleet-metadata=%s", iso.FleetMetadata),
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo /home/core/bin/gluon setup %s", strings.Join(gluonArgs, " ")), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// UpdateClusterMembers updates /etc/pulcy/cluster-members on the given instance
func (i ClusterInstance) UpdateClusterMembers(log *logging.Logger, members ClusterMemberList) error {
	if _, err := i.runRemoteCommand(log, "sudo /usr/bin/mkdir -p /etc/pulcy", "", false); err != nil {
		return maskAny(err)
	}
	data := members.Render()
	if _, err := i.runRemoteCommand(log, "sudo tee /etc/pulcy/cluster-members", data, false); err != nil {
		return maskAny(err)
	}

	log.Info("Restarting gluon on %s", i.PublicIpv4)
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo systemctl restart gluon.service"), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}
