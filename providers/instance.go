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
	"path"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/juju/errgo"
	"github.com/op/go-logging"
)

const (
	defaultUsername = "core"
)

// User returns the standard username of this instance
func (i ClusterInstance) User() string {
	if i.UserName == "" {
		return defaultUsername
	}
	return i.UserName
}

// User returns the standard home directory instance
func (i ClusterInstance) Home() string {
	switch i.User() {
	case "root":
		return "/root"
	default:
		return fmt.Sprintf("/home/%s", i.User())
	}
}

func (i ClusterInstance) runRemoteCommand(log *logging.Logger, command, stdin string, quiet bool) (string, error) {
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", i.User()+"@"+i.PublicIpv4, command)
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
	log.Debugf("Fetching cluster-id on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "sudo cat /etc/pulcy/cluster-id", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetMachineID(log *logging.Logger) (string, error) {
	log.Debugf("Fetching machine-id on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "cat /etc/machine-id", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetVaultCrt(log *logging.Logger) (string, error) {
	log.Debugf("Fetching vault.crt on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "sudo cat /etc/pulcy/vault.crt", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetVaultAddr(log *logging.Logger) (string, error) {
	const prefix = "VAULT_ADDR="
	log.Debugf("Fetching vault-addr on %s", i.PublicIpv4)
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

func (i ClusterInstance) GetOSRelease(log *logging.Logger) (semver.Version, error) {
	const prefix = "DISTRIB_RELEASE="
	log.Debugf("Fetching OS release on %s", i.PublicIpv4)
	env, err := i.runRemoteCommand(log, "cat /etc/lsb-release", "", false)
	if err != nil {
		return semver.Version{}, maskAny(err)
	}
	for _, line := range strings.Split(env, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			v, err := semver.NewVersion(strings.TrimSpace(line[len(prefix):]))
			if err != nil {
				return semver.Version{}, maskAny(err)
			}
			return *v, nil
		}
	}
	return semver.Version{}, maskAny(errgo.Newf("%s not found in /etc/lsb-release", prefix))
}

func (i ClusterInstance) IsEtcdProxy(log *logging.Logger) (bool, error) {
	log.Debugf("Fetching etcd proxy status on %s", i.PublicIpv4)
	cat, err := i.runRemoteCommand(log, "systemctl cat etcd2.service", "", false)
	return strings.Contains(cat, "ETCD_PROXY"), maskAny(err)
}

// AddEtcdMember calls etcdctl to add a member to ETCD
func (i ClusterInstance) AddEtcdMember(log *logging.Logger, name, privateIP string) error {
	log.Infof("Adding %s(%s) to etcd on %s", name, privateIP, i.PublicIpv4)
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
	log.Infof("Removing %s(%s) from etcd on %s", name, privateIP, i.PublicIpv4)
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

func (i ClusterInstance) waitUntilActive(log *logging.Logger) error {
	for {
		// Attempt an SSH connection
		if _, err := i.GetMachineID(log); err == nil {
			// Success
			return nil
		}
		// Wait a while
		time.Sleep(time.Second * 5)
	}
}

// osSetup updates the OS of the instance (if needed)
func (i ClusterInstance) osSetup(log *logging.Logger, minOSVersion semver.Version) error {
	v, err := i.GetOSRelease(log)
	if err != nil {
		return maskAny(err)
	}
	if !v.LessThan(minOSVersion) {
		// OS is up to date
		log.Infof("OS on %s is up to date", i.PublicIpv4)
		return nil
	}
	// Run update
	log.Infof("Updating OS on %s...", i.PublicIpv4)
	if _, err := i.runRemoteCommand(log, "sudo update_engine_client -update", "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo -b shutdown -r now", "", false); err != nil {
		// This will likely fail
		log.Debugf("Reboot failed (likely): %#v", err)
	}
	time.Sleep(time.Second * 5)
	// Wait until available
	if err := i.waitUntilActive(log); err != nil {
		return maskAny(err)
	}
	return nil
}

// InitialSetup creates initial files and calls gluon for the first time
func (i ClusterInstance) InitialSetup(log *logging.Logger, cio CreateInstanceOptions, iso InitialSetupOptions) error {
	if !i.NoCoreOS {
		minOSVersion, err := semver.NewVersion(cio.MinOSVersion)
		if err != nil {
			return maskAny(err)
		}
		if err := i.osSetup(log, *minOSVersion); err != nil {
			return maskAny(err)
		}
	}

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

	log.Infof("Downloading gluon on %s", i.PublicIpv4)
	binDir := path.Join(i.Home(), "bin")
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo /usr/bin/mkdir -p %s", binDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("docker run --rm -v %s:/destination/ %s", binDir, cio.GluonImage), "", false); err != nil {
		return maskAny(err)
	}
	log.Infof("Running gluon on %s", i.PublicIpv4)
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
	gluonPath := path.Join(binDir, "gluon")
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo %s setup %s", gluonPath, strings.Join(gluonArgs, " ")), "", false); err != nil {
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

	log.Infof("Restarting gluon on %s", i.PublicIpv4)
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo systemctl restart gluon.service"), "", false); err != nil {
		return maskAny(err)
	}

	log.Infof("Enabling services on %s", i.PublicIpv4)
	services := []string{"etcd2.service", "fleet.service", "fleet.socket"}
	for _, service := range services {
		if err := i.EnableService(log, service); err != nil {
			return maskAny(err)
		}
	}
	return nil
}

// Reboot reboots the instance
func (i ClusterInstance) Reboot(log *logging.Logger) error {
	if _, err := i.runRemoteCommand(log, "sudo sync", "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo reboot -f", "", false); err != nil {
		// This will likely fail
		log.Debugf("Reboot failed (likely): %#v", err)
	}
	return nil
}

// EnableService calls `systemctl enable <name>`
func (i ClusterInstance) EnableService(log *logging.Logger, name string) error {
	if _, err := i.runRemoteCommand(log, "sudo systemctl enable "+name, "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// RunScript uploads a script with given content and executes it
func (i ClusterInstance) RunScript(log *logging.Logger, scriptContent, scriptPath string) error {
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo tee %s", scriptPath), scriptContent, false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo chmod +x %s", scriptPath), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo %s", scriptPath), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}
