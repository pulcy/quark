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
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/coreos/go-semver/semver"
	"github.com/op/go-logging"
)

const (
	defaultUsername = "core"
	sshPort         = 22
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
	ClusterIP        string // IPv4 address of the instance used for all private communication in the cluster
	LoadBalancerIPv4 string // IPv4 address of the instance on which the load-balancer is listening (can be empty)
	LoadBalancerIPv6 string // IPv6 address of the instance on which the load-balancer is listening (can be empty)
	IsGateway        bool   // If set, this instance can be used as a gateway by instances that have not direct IPv4 internet connection
	LoadBalancerDNS  string // Provider hosted public DNS name of the instance on which the load-balancer is listening (can be empty)
	ClusterDevice    string // Device name of the nic that is configured for the ClusterIP
	PrivateIP        string // IP address of the instance's private network (can be same as ClusterIP)
	PrivateNetwork   net.IPNet
	PrivateDNS       string   // Provider hosted private DNS name of the instance's private network
	UserName         string   // Account name used to SSH into this instance. (empty defaults to 'core')
	OS               OSName   // Name of the OS on the instance
	Extra            []string // Extra informational data
	EtcdProxy        *bool
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
	if i.LoadBalancerDNS != "" {
		return i.LoadBalancerDNS
	}
	return i.Name
}

func (i ClusterInstance) hostAddress(addBrackets bool) string {
	s := i.String()
	if addBrackets && strings.Contains(s, ":") {
		return "[" + s + "]"
	}
	return s
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

// IsSSHPortOpen checks if the SSH port on this instance is open for communications.
func (i ClusterInstance) IsSSHPortOpen(log *logging.Logger) (bool, error) {
	log.Debugf("Testing SSH port status on %s", i)
	hostAddress := i.String()
	if hostAddress == "" {
		return false, maskAny(fmt.Errorf("don't have any address to communicate with instance %s", i.Name))
	}
	return isTCPPortOpen(hostAddress, sshPort), nil
}

// Connect opens an SSH session to the instance.
// Make sure to close the session when done.
func (i ClusterInstance) Connect() (InstanceConnection, error) {
	hostAddress := i.String()
	if hostAddress == "" {
		return nil, maskAny(fmt.Errorf("don't have any address to communicate with instance %s", i.Name))
	}
	client, err := SSHConnect(i.User(), hostAddress)
	if err != nil {
		return nil, maskAny(err)
	}
	return client, nil
}

// GetMachineID loads the machine specific unique ID of the instance.
func (i ClusterInstance) GetMachineID(log *logging.Logger) (string, error) {
	s, err := i.Connect()
	if err != nil {
		return "", maskAny(err)
	}
	defer s.Close()
	id, err := s.GetMachineID(log)
	if err != nil {
		return "", maskAny(err)
	}
	return id, nil
}

// IsEtcdProxy returns true if the instance in an ETCD proxy.
func (i ClusterInstance) IsEtcdProxy(log *logging.Logger) (bool, error) {
	if i.EtcdProxy != nil {
		return *i.EtcdProxy, nil
	}
	s, err := i.Connect()
	if err != nil {
		return false, maskAny(err)
	}
	defer s.Close()
	result, err := s.IsEtcdProxyFromService(log)
	if err != nil {
		return false, maskAny(err)
	}
	return result, nil
}

// AsClusterMember fetches all data from the instance needed for a ClusterMember and returns that.
func (i ClusterInstance) AsClusterMember(log *logging.Logger) (ClusterMember, error) {
	result := ClusterMember{
		ClusterIP:     i.ClusterIP,
		PrivateHostIP: i.PrivateIP,
	}
	s, err := i.Connect()
	if err != nil {
		return ClusterMember{}, maskAny(err)
	}
	defer s.Close()

	clusterID, err := s.GetClusterID(log)
	if err != nil {
		return ClusterMember{}, maskAny(err)
	}
	result.ClusterID = clusterID

	machineID, err := s.GetMachineID(log)
	if err != nil {
		return ClusterMember{}, maskAny(err)
	}
	result.MachineID = machineID

	etcdProxy, err := i.IsEtcdProxy(log)
	if err != nil {
		return ClusterMember{}, maskAny(err)
	}
	result.EtcdProxy = etcdProxy

	return result, nil
}

type InitialSetupOptions struct {
	ClusterMembers   ClusterMemberList
	FleetMetadata    string
	EtcdClusterState string
}

// waitUntilActive blocks until the instance is alive and its machine ID can be fetched.
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

// waitUntilInternetConnection blocks until the instance can ping to 8.8.8.8.
func (i ClusterInstance) waitUntilInternetConnection(log *logging.Logger) error {
	for {
		// Attempt an SSH connection
		if s, err := i.Connect(); err == nil {
			if _, err := s.Run(log, "ping -c 3 -w 60 8.8.8.8", "", true); err == nil {
				// Success
				s.Close()
				return nil
			}
			s.Close()
		}
		// Wait a while
		time.Sleep(time.Second * 2)
	}
}

// osSetup updates the OS of the instance (if needed)
func (i ClusterInstance) osSetup(s InstanceConnection, log *logging.Logger, minOSVersion semver.Version, provider CloudProvider) error {
	v, err := s.GetOSRelease(log)
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
	if _, err := s.Run(log, "sudo update_engine_client -update", "", false); err != nil {
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
	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()

	if i.OS == OSNameCoreOS {
		minOSVersion, err := semver.NewVersion(cio.MinOSVersion)
		if err != nil {
			return maskAny(err)
		}
		if err := i.osSetup(s, log, *minOSVersion, provider); err != nil {
			return maskAny(err)
		}
	}

	if _, err := s.Run(log, "sudo /usr/bin/mkdir -p /etc/pulcy", "", false); err != nil {
		return maskAny(err)
	}
	data := iso.ClusterMembers.Render()
	if _, err := s.Run(log, "sudo tee /etc/pulcy/cluster-members", data, false); err != nil {
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
	var vaultServerKey string
	if cio.RoleVault {
		vaultServerKey, err = cio.VaultServerKey()
		if err != nil {
			return maskAny(err)
		}
	}
	if _, err := s.Run(log, "sudo tee /etc/pulcy/vault.env", strings.Join(vaultEnv, "\n"), false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, "sudo chmod 0400 /etc/pulcy/vault.env", "", false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, "sudo tee /etc/pulcy/vault.crt", vaultCertificate, false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, "sudo chmod 0400 /etc/pulcy/vault.crt", "", false); err != nil {
		return maskAny(err)
	}
	if cio.RoleVault {
		if _, err := s.Run(log, "sudo /usr/bin/mkdir -p /etc/pulcy/vault", "", false); err != nil {
			return maskAny(err)
		}
		if _, err := s.Run(log, "sudo tee /etc/pulcy/vault/key.pem", vaultServerKey, false); err != nil {
			return maskAny(err)
		}
		if _, err := s.Run(log, "sudo chmod 0400 /etc/pulcy/vault/key.pem", "", false); err != nil {
			return maskAny(err)
		}
	}

	if _, err := s.Run(log, "sudo tee /etc/pulcy/gluon.env", cio.GluonEnv, false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, "sudo chmod 0644 /etc/pulcy/gluon.env", "", false); err != nil {
		return maskAny(err)
	}

	if _, err := s.Run(log, "sudo tee /etc/pulcy/weave.env", cio.WeaveEnv, false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, "sudo chmod 0400 /etc/pulcy/weave.env", "", false); err != nil {
		return maskAny(err)
	}
	if cio.WeaveSeed != "" {
		if _, err := s.Run(log, "sudo tee /etc/pulcy/weave-seed", cio.WeaveSeed, false); err != nil {
			return maskAny(err)
		}
	}

	log.Infof("Waiting for internet connection on %s", i)
	if err := i.waitUntilInternetConnection(log); err != nil {
		return maskAny(err)
	}

	log.Infof("Downloading gluon on %s", i)
	binDir := path.Join(i.Home(), "bin")
	if _, err := s.Run(log, fmt.Sprintf("sudo /usr/bin/mkdir -p %s", binDir), "", false); err != nil {
		return maskAny(err)
	}
	// Docker registry is not always stable to retry if needed
	extractGluon := func() error {
		if _, err := s.Run(log, fmt.Sprintf("docker run --rm -v %s:/destination/ %s", binDir, cio.GluonImage), "", false); err != nil {
			log.Warningf("Extracting gluon failed: %#v", err)
			return maskAny(err)
		}
		return nil
	}
	if err := backoff.Retry(extractGluon, backoff.NewExponentialBackOff()); err != nil {
		return maskAny(err)
	}

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
	log.Infof("Running gluon on %s", i)
	log.Debugf("Gluon args on %s: %#v", i, gluonArgs)
	gluonPath := path.Join(binDir, "gluon")
	if _, err := s.Run(log, fmt.Sprintf("sudo %s setup %s", gluonPath, strings.Join(gluonArgs, " ")), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// UpdateClusterMembers updates /etc/pulcy/cluster-members on the given instance
func (i ClusterInstance) UpdateClusterMembers(log *logging.Logger, members ClusterMemberList) error {
	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()

	if _, err := s.Run(log, "sudo /usr/bin/mkdir -p /etc/pulcy", "", false); err != nil {
		return maskAny(err)
	}
	data := members.Render()
	if _, err := s.Run(log, "sudo tee /etc/pulcy/cluster-members", data, false); err != nil {
		return maskAny(err)
	}

	log.Infof("Restarting gluon on %s", i)
	if _, err := s.Run(log, fmt.Sprintf("sudo systemctl restart gluon.service"), "", false); err != nil {
		return maskAny(err)
	}

	log.Infof("Enabling services on %s", i)
	services := []string{"ip4tables.service", "ip6tables.service"}
	for _, service := range services {
		if err := s.EnableService(log, service); err != nil {
			return maskAny(err)
		}
	}
	return nil
}

// isTCPPortOpen returns true if a TCP communication with "host:port" can be initialized
func isTCPPortOpen(host string, port int) bool {
	dest := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", dest, time.Duration(2000)*time.Millisecond)
	if err == nil {
		defer conn.Close()
	}
	return err == nil
}
