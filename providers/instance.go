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

// OSName specifies a name of an OS
type OSName string

const (
	OSNameCoreOS OSName = "coreos"
	OSNameUbuntu OSName = "ubuntu"
)

// ClusterInstance describes a single instance
type ClusterInstance struct {
	ID               string // Provider specific ID of the server (only used by provider, can be empty)
	Name             string // Name of the instance as known by the provider
	ClusterIP        string // IP address of the instance used for all private communication in the cluster
	LoadBalancerIPv4 string // IPv4 address of the instance on which the load-balancer is listening (can be empty)
	LoadBalancerIPv6 string // IPv6 address of the instance on which the load-balancer is listening (can be empty)
	ClusterDevice    string // Device name of the nic that is configured for the ClusterIP
	PrivateIP        string // IP address of the instance's private network (can be same as ClusterIP)
	UserName         string // Account name used to SSH into this instance. (empty defaults to 'core')
	OS               OSName // Name of the OS on the instance
}

// Equals returns true of the given cluster instances refer to the same instance.
func (i ClusterInstance) Equals(other ClusterInstance) bool {
	return i.ID == other.ID && i.ClusterIP == other.ClusterIP
}

// String returns a human readable representation of the given instance
func (i ClusterInstance) String() string {
	if i.LoadBalancerIPv4 != "" {
		return i.LoadBalancerIPv4
	}
	if i.LoadBalancerIPv6 != "" {
		return i.LoadBalancerIPv6
	}
	return i.Name
}

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
	hostAddress := i.LoadBalancerIPv4
	if hostAddress == "" {
		hostAddress = i.LoadBalancerIPv6
	}
	if hostAddress == "" {
		return "", maskAny(fmt.Errorf("don't have any address to communicate with instance %s", i.Name))
	}
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", i.User()+"@"+hostAddress, command)
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
	log.Debugf("Fetching cluster-id on %s", i)
	id, err := i.runRemoteCommand(log, "sudo cat /etc/pulcy/cluster-id", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetMachineID(log *logging.Logger) (string, error) {
	log.Debugf("Fetching machine-id on %s", i)
	id, err := i.runRemoteCommand(log, "cat /etc/machine-id", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetVaultCrt(log *logging.Logger) (string, error) {
	log.Debugf("Fetching vault.crt on %s", i)
	id, err := i.runRemoteCommand(log, "sudo cat /etc/pulcy/vault.crt", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetVaultAddr(log *logging.Logger) (string, error) {
	const prefix = "VAULT_ADDR="
	log.Debugf("Fetching vault-addr on %s", i)
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
	log.Debugf("Fetching OS release on %s", i)
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
	log.Debugf("Fetching etcd proxy status on %s", i)
	cat, err := i.runRemoteCommand(log, "systemctl cat etcd2.service", "", false)
	return strings.Contains(cat, "ETCD_PROXY"), maskAny(err)
}

// AddEtcdMember calls etcdctl to add a member to ETCD
func (i ClusterInstance) AddEtcdMember(log *logging.Logger, name, clusterIP string) error {
	log.Infof("Adding %s(%s) to etcd on %s", name, clusterIP, i)
	cmd := []string{
		"etcdctl",
		"member",
		"add",
		name,
		fmt.Sprintf("http://%s:2380", clusterIP),
	}
	if _, err := i.runRemoteCommand(log, strings.Join(cmd, " "), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// RemoveEtcdMember calls etcdctl to remove a member from ETCD
func (i ClusterInstance) RemoveEtcdMember(log *logging.Logger, name, clusterIP string) error {
	log.Infof("Removing %s(%s) from etcd on %s", name, clusterIP, i)
	id, err := i.runRemoteCommand(log, fmt.Sprintf("sh -c 'etcdctl member list | grep %s | cut -d: -f1 | cut -d[ -f1'", clusterIP), "", false)
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
		ClusterIP: i.ClusterIP,
		EtcdProxy: etcdProxy,
	}, nil
}

type InitialSetupOptions struct {
	ClusterMembers   ClusterMemberList
	FleetMetadata    string
	EtcdClusterState string
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
func (i ClusterInstance) osSetup(log *logging.Logger, minOSVersion semver.Version, provider CloudProvider) error {
	v, err := i.GetOSRelease(log)
	if err != nil {
		return maskAny(err)
	}
	if !v.LessThan(minOSVersion) {
		// OS is up to date
		log.Infof("OS on %s is up to date", i)
		return nil
	}
	// Run update
	log.Infof("Updating OS on %s...", i)
	if _, err := i.runRemoteCommand(log, "sudo update_engine_client -update", "", false); err != nil {
		return maskAny(err)
	}
	if err := provider.RebootInstance(i); err != nil {
		// This may likely fail
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
func (i ClusterInstance) InitialSetup(log *logging.Logger, cio CreateInstanceOptions, iso InitialSetupOptions, provider CloudProvider) error {
	if i.OS == OSNameCoreOS {
		minOSVersion, err := semver.NewVersion(cio.MinOSVersion)
		if err != nil {
			return maskAny(err)
		}
		if err := i.osSetup(log, *minOSVersion, provider); err != nil {
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
	vaultCertificate, err := cio.VaultCertificate()
	if err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo tee /etc/pulcy/vault.env", strings.Join(vaultEnv, "\n"), false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo chmod 0400 /etc/pulcy/vault.env", "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo tee /etc/pulcy/vault.crt", vaultCertificate, false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo chmod 0400 /etc/pulcy/vault.crt", "", false); err != nil {
		return maskAny(err)
	}

	log.Infof("Downloading gluon on %s", i)
	binDir := path.Join(i.Home(), "bin")
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo /usr/bin/mkdir -p %s", binDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("docker run --rm -v %s:/destination/ %s", binDir, cio.GluonImage), "", false); err != nil {
		return maskAny(err)
	}
	log.Infof("Running gluon on %s", i)
	gluonArgs := []string{
		fmt.Sprintf("--gluon-image=%s", cio.GluonImage),
		fmt.Sprintf("--docker-ip=%s", i.ClusterIP),
		fmt.Sprintf("--private-ip=%s", i.ClusterIP),
		fmt.Sprintf("--private-cluster-device=%s", i.ClusterDevice),
		fmt.Sprintf("--private-registry-url=%s", cio.PrivateRegistryUrl),
		fmt.Sprintf("--private-registry-username=%s", cio.PrivateRegistryUserName),
		fmt.Sprintf("--private-registry-password=%s", cio.PrivateRegistryPassword),
		fmt.Sprintf("--fleet-metadata=%s", iso.FleetMetadata),
	}
	if iso.EtcdClusterState != "" {
		gluonArgs = append(gluonArgs, fmt.Sprintf("--etcd-cluster-state=%s", iso.EtcdClusterState))
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

	log.Infof("Restarting gluon on %s", i)
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo systemctl restart gluon.service"), "", false); err != nil {
		return maskAny(err)
	}

	log.Infof("Enabling services on %s", i)
	services := []string{"etcd2.service", "fleet.service", "fleet.socket", "ip4tables.service", "ip6tables.service"}
	for _, service := range services {
		if err := i.EnableService(log, service); err != nil {
			return maskAny(err)
		}
	}
	return nil
}

// Sync the filesystems on the instance
func (i ClusterInstance) Sync(log *logging.Logger) error {
	if _, err := i.runRemoteCommand(log, "sudo sync", "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// Exec executes a command on the instance
func (i ClusterInstance) Exec(log *logging.Logger, command string) (string, error) {
	stdout, err := i.runRemoteCommand(log, command, "", false)
	if err != nil {
		return stdout, maskAny(err)
	}
	return stdout, nil
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
