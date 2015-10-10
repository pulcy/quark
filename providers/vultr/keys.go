package vultr

import (
	"fmt"
	"sort"

	"github.com/juju/errgo"
	"github.com/ryanuber/columnize"
)

func (vp *vultrProvider) ShowKeys() error {
	keys, err := vp.client.GetSSHKeys()
	if err != nil {
		return maskAny(err)
	}
	lines := []string{
		"ID | Name | Public-key",
	}
	for _, r := range keys {
		line := fmt.Sprintf("%v | %s | %s", r.ID, r.Name, r.Key)
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}

// Search for an SSH key with given name and return its ID
func (vp *vultrProvider) findSSHKeyID(keyName string) (string, error) {
	keys, err := vp.client.GetSSHKeys()
	if err != nil {
		return "", maskAny(err)
	}
	for _, k := range keys {
		if k.Name == keyName {
			return k.ID, nil
		}
	}
	return "", errgo.WithCausef(nil, InvalidArgumentError, "key %s not found", keyName)
}
