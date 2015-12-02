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
			log.Error("SSH failed: %s %s", cmd.Path, strings.Join(cmd.Args, " "))
		}
		return "", errgo.NoteMask(err, stdErr.String())
	}

	out := stdOut.String()
	out = strings.TrimSuffix(out, "\n")
	return out, nil
}

func (i ClusterInstance) GetMachineID(log *logging.Logger) (string, error) {
	log.Info("Fetching machine-id on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "cat /etc/machine-id", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) GetEtcdDiscoveryURL(log *logging.Logger) (string, error) {
	log.Info("Fetching discovery-url on %s", i.PublicIpv4)
	uuid, err := i.runRemoteCommand(log, "cat /etc/pulcy/discovery-url", "", true)
	if err != nil {
		uuid, err = i.runRemoteCommand(log, "systemctl cat etcd2 | grep ETCD_DISCOVERY |grep -oE 'http.*/[a-z0-9A-Z]+'", "", false)
	}
	return uuid, maskAny(err)
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

// UpdateClusterMembers updates /etc/yard-cluster-members on the given instance
func (i ClusterInstance) UpdateClusterMembers(log *logging.Logger, clusterIPs []string) error {
	data := strings.Join(clusterIPs, "\n")
	if _, err := i.runRemoteCommand(log, "sudo tee /etc/yard-cluster-members", data, false); err != nil {
		return maskAny(err)
	}
	discoveryURL, err := i.GetEtcdDiscoveryURL(log)
	if err != nil {
		return maskAny(err)
	}
	log.Info("Updating cluster iptable rules on %s", i.PublicIpv4)
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo /home/core/bin/yard cluster update --discovery-url %s", discoveryURL), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}
