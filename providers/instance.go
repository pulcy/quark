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
	log.Info("Fetching cluster-id on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "sudo cat /etc/pulcy/cluster-id", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetMachineID(log *logging.Logger) (string, error) {
	log.Info("Fetching machine-id on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "cat /etc/machine-id", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetVaultCrt(log *logging.Logger) (string, error) {
	log.Info("Fetching vault.crt on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "sudo cat /etc/pulcy/vault.crt", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetVaultAddr(log *logging.Logger) (string, error) {
	const prefix = "VAULT_ADDR="
	log.Info("Fetching vault-addr on %s", i.PublicIpv4)
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
	log.Info("Fetching etcd proxy status on %s", i.PublicIpv4)
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

type ClusterMember struct {
	MachineID string
	PrivateIP string
	EtcdProxy bool
}

// UpdateClusterMembers updates /etc/pulcy/cluster-members on the given instance
func (i ClusterInstance) UpdateClusterMembers(log *logging.Logger, members []ClusterMember) error {
	data := ""
	for _, cm := range members {
		proxy := ""
		if cm.EtcdProxy {
			proxy = " etcd-proxy"
		}
		data = data + fmt.Sprintf("%s=%s%s\n", cm.MachineID, cm.PrivateIP, proxy)
	}
	if _, err := i.runRemoteCommand(log, "sudo /usr/bin/mkdir -p /etc/pulcy", "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo tee /etc/pulcy/cluster-members", data, false); err != nil {
		return maskAny(err)
	}
	log.Info("Restarting gluon on %s", i.PublicIpv4)
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo systemctl restart gluon.service"), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}
