package providers

import (
	"bytes"
	"io"
	"net"
	"os"
	"strings"

	"github.com/juju/errgo"
	logging "github.com/op/go-logging"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHClient interface {
	io.Closer
	Run(log *logging.Logger, command, stdin string, quiet bool) (string, error)
}

type sshClient struct {
	client *ssh.Client
}

// DialSSH creates a new SSH connection to the given user on the given host.
func DialSSH(userName, host string) (SSHClient, error) {
	// To authenticate with the remote server you must pass at least one
	// implementation of AuthMethod via the Auth field in ClientConfig.
	config := &ssh.ClientConfig{
		User: userName,
	}
	var sshAgent agent.Agent
	if agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		sshAgent = agent.NewClient(agentConn)
		config.Auth = append(config.Auth, ssh.PublicKeysCallback(sshAgent.Signers))
	} else {
		return nil, maskAny(err)
	}

	addr := net.JoinHostPort(host, "22")
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, maskAny(err)
	}

	return &sshClient{client}, nil
}

func (s *sshClient) Close() error {
	return maskAny(s.client.Close())
}

func (s *sshClient) Run(log *logging.Logger, command, stdin string, quiet bool) (string, error) {
	var stdOut, stdErr bytes.Buffer

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := s.client.NewSession()
	if err != nil {
		return "", maskAny(err)
	}
	defer session.Close()

	session.Stdout = &stdOut
	session.Stderr = &stdErr

	if stdin != "" {
		session.Stdin = strings.NewReader(stdin)
	}

	if err := session.Run(command); err != nil {
		if !quiet {
			log.Errorf("SSH failed: %s", command)
		}
		return "", errgo.NoteMask(err, stdErr.String())
	}

	out := stdOut.String()
	out = strings.TrimSuffix(out, "\n")
	return out, nil
}
