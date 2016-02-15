package vagrant

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

const (
	insecurePrivateKeyPathTmpl = "~/.vagrant.d/insecure_private_key"
)

func fetchVagrantInsecureSSHKey() (string, error) {
	insecurePrivateKeyPath, err := homedir.Expand(insecurePrivateKeyPathTmpl)
	if err != nil {
		return "", maskAny(err)
	}

	privateRaw, err := ioutil.ReadFile(insecurePrivateKeyPath)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", maskAny(err)
	}

	privateKey, err := ssh.ParsePrivateKey(privateRaw)
	if err != nil {
		return "", maskAny(err)
	}

	resultRaw := ssh.MarshalAuthorizedKey(privateKey.PublicKey())
	return strings.TrimSpace(string(resultRaw)), nil
}
