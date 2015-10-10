package vultr

import (
	"fmt"
	"sort"

	"github.com/ryanuber/columnize"
)

func (vp *vultrProvider) ShowImages() error {
	// Load OS's
	os, err := vp.client.GetOS()
	if err != nil {
		return maskAny(err)
	}

	lines := []string{
		"ID | Name | Arch | Family",
	}
	for _, r := range os {
		line := fmt.Sprintf("%d | %s | %s | %s", r.ID, r.Name, r.Arch, r.Family)
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}
