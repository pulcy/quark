package vultr

import (
	"fmt"
	"sort"

	"github.com/ryanuber/columnize"
)

func (vp *vultrProvider) ShowRegions() error {
	regions, err := vp.client.GetRegions()
	if err != nil {
		return maskAny(err)
	}

	lines := []string{
		"ID | Name | State | Country | Continent",
	}
	for _, r := range regions {
		line := fmt.Sprintf("%02d | %s | %s | %s | %s", r.ID, r.Name, r.State, r.Country, r.Continent)
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}
