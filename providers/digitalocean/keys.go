package digitalocean

import (
	"fmt"
	"sort"

	"github.com/ryanuber/columnize"
)

func (this *doProvider) ShowKeys() error {
	// Load images
	client := NewDOClient(this.token)
	keys, err := KeyList(client)
	if err != nil {
		return err
	}

	lines := []string{
		"ID | Name | FingerPrint | Public-key",
	}
	for _, r := range keys {
		line := fmt.Sprintf("%v | %s | %s | %s", r.ID, r.Name, r.Fingerprint, r.PublicKey)
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}
