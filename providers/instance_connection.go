package providers

import (
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/juju/errgo"
	logging "github.com/op/go-logging"
)

type InstanceConnection interface {
	SSHClient

	// Sync the filesystems on the instance
	Sync(log *logging.Logger) error

	// Exec executes a command on the instance
	Exec(log *logging.Logger, command string) (string, error)

	// EnableService calls `systemctl enable <name>`
	EnableService(log *logging.Logger, name string) error
	// RunScript uploads a script with given content and executes it
	RunScript(log *logging.Logger, scriptContent, scriptPath string) error

	GetClusterID(log *logging.Logger) (string, error)

	GetGluonEnv(log *logging.Logger) (string, error)

	GetMachineID(log *logging.Logger) (string, error)

	GetVaultCrt(log *logging.Logger) (string, error)

	GetVaultAddr(log *logging.Logger) (string, error)

	GetWeaveEnv(log *logging.Logger) (string, error)

	GetWeaveSeed(log *logging.Logger) (string, error)

	GetOSRelease(log *logging.Logger) (semver.Version, error)

	// IsEtcdProxyFromService queries the ETCD2 service on the instance to look for an ETCD_PROXY variable.
	IsEtcdProxyFromService(log *logging.Logger) (bool, error)

	// AddEtcdMember calls etcdctl to add a member to ETCD
	AddEtcdMember(log *logging.Logger, name, clusterIP string) error

	// RemoveEtcdMember calls etcdctl to remove a member from ETCD
	RemoveEtcdMember(log *logging.Logger, name, clusterIP string) error
}

type instanceConnection struct {
	client SSHClient
	host   string
}

// SSHConnect creates a new SSH connection to the given user on the given host.
func SSHConnect(userName, host string) (InstanceConnection, error) {
	client, err := DialSSH(userName, host)
	if err != nil {
		return nil, maskAny(err)
	}
	return &instanceConnection{client, host}, nil
}

func (s *instanceConnection) Close() error {
	return maskAny(s.client.Close())
}

func (s *instanceConnection) Run(log *logging.Logger, command, stdin string, quiet bool) (string, error) {
	out, err := s.client.Run(log, command, stdin, quiet)
	return out, maskAny(err)
}

// Sync the filesystems on the instance
func (s *instanceConnection) Sync(log *logging.Logger) error {
	if _, err := s.Run(log, "sudo sync", "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// Exec executes a command on the instance
func (s *instanceConnection) Exec(log *logging.Logger, command string) (string, error) {
	stdout, err := s.Run(log, command, "", false)
	if err != nil {
		return stdout, maskAny(err)
	}
	return stdout, nil
}

// EnableService calls `systemctl enable <name>`
func (s *instanceConnection) EnableService(log *logging.Logger, name string) error {
	if _, err := s.Run(log, "sudo systemctl enable "+name, "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// RunScript uploads a script with given content and executes it
func (s *instanceConnection) RunScript(log *logging.Logger, scriptContent, scriptPath string) error {
	if _, err := s.Run(log, fmt.Sprintf("sudo tee %s", scriptPath), scriptContent, false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo chmod +x %s", scriptPath), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo %s", scriptPath), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

func (s *instanceConnection) GetClusterID(log *logging.Logger) (string, error) {
	log.Debugf("Fetching cluster-id on %s", s.host)
	id, err := s.Run(log, "sudo cat /etc/pulcy/cluster-id", "", false)
	return id, maskAny(err)
}

func (s *instanceConnection) GetGluonEnv(log *logging.Logger) (string, error) {
	log.Debugf("Fetching gluon.env on %s", s.host)
	// gluon.env does not have to exists, so ignore errors by the `|| echo ""` parts.
	id, err := s.Run(log, "sh -c 'sudo cat /etc/pulcy/gluon.env || echo \"\"'", "", false)
	return id, maskAny(err)
}

func (s *instanceConnection) GetMachineID(log *logging.Logger) (string, error) {
	log.Debugf("Fetching machine-id on %s", s.host)
	id, err := s.Run(log, "cat /etc/machine-id", "", false)
	return id, maskAny(err)
}

func (s *instanceConnection) GetVaultCrt(log *logging.Logger) (string, error) {
	log.Debugf("Fetching vault.crt on %s", s.host)
	id, err := s.Run(log, "sudo cat /etc/pulcy/vault.crt", "", false)
	return id, maskAny(err)
}

func (s *instanceConnection) GetVaultAddr(log *logging.Logger) (string, error) {
	const prefix = "VAULT_ADDR="
	log.Debugf("Fetching vault-addr on %s", s.host)
	env, err := s.Run(log, "sudo cat /etc/pulcy/vault.env", "", false)
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

func (s *instanceConnection) GetWeaveEnv(log *logging.Logger) (string, error) {
	log.Debugf("Fetching weave.env on %s", s.host)
	id, err := s.Run(log, "sudo cat /etc/pulcy/weave.env", "", false)
	return id, maskAny(err)
}

func (s *instanceConnection) GetWeaveSeed(log *logging.Logger) (string, error) {
	log.Debugf("Fetching weave-seed on %s", s.host)
	// weave-seed does not have to exists, so ignore errors by the `|| echo ""` parts.
	id, err := s.Run(log, "sh -c 'sudo cat /etc/pulcy/weave-seed || echo \"\"'", "", false)
	return id, maskAny(err)
}

func (s *instanceConnection) GetOSRelease(log *logging.Logger) (semver.Version, error) {
	const prefix = "DISTRIB_RELEASE="
	log.Debugf("Fetching OS release on %s", s.host)
	env, err := s.Run(log, "cat /etc/lsb-release", "", false)
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

// IsEtcdProxyFromService queries the ETCD2 service on the instance to look for an ETCD_PROXY variable.
func (s *instanceConnection) IsEtcdProxyFromService(log *logging.Logger) (bool, error) {
	log.Debugf("Fetching etcd proxy status on %s", s.host)
	cat, err := s.Run(log, "sudo systemctl cat etcd2.service", "", false)
	return strings.Contains(cat, "ETCD_PROXY"), maskAny(err)
}

// AddEtcdMember calls etcdctl to add a member to ETCD
func (s *instanceConnection) AddEtcdMember(log *logging.Logger, name, clusterIP string) error {
	log.Infof("Adding %s(%s) to etcd on %s", name, clusterIP, s.host)
	cmd := []string{
		"etcdctl",
		"member",
		"add",
		name,
		fmt.Sprintf("http://%s:2380", clusterIP),
	}
	if _, err := s.Run(log, strings.Join(cmd, " "), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// RemoveEtcdMember calls etcdctl to remove a member from ETCD
func (s *instanceConnection) RemoveEtcdMember(log *logging.Logger, name, clusterIP string) error {
	log.Infof("Removing %s(%s) from etcd on %s", name, clusterIP, s.host)
	id, err := s.Run(log, fmt.Sprintf("sh -c 'etcdctl member list | grep %s | cut -d: -f1 | cut -d[ -f1'", clusterIP), "", false)
	if err != nil {
		return maskAny(err)
	}
	cmd := []string{
		"etcdctl",
		"member",
		"remove",
		id,
	}
	if _, err := s.Run(log, strings.Join(cmd, " "), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}
