package vultr

import (
	"fmt"
	"sort"

	"github.com/ryanuber/columnize"
)

func (vp *vultrProvider) ShowInstanceTypes() error {
	plans, err := vp.client.GetPlans()
	if err != nil {
		return maskAny(err)
	}

	lines := []string{
		"ID | Name | VCpu | RAM | Disk | Bandwidth | Price",
	}
	for _, p := range plans {
		line := fmt.Sprintf("%03d | %s | %d | %s | %s | %s | %s", p.ID, p.Name, p.VCpus, p.RAM, p.Disk, p.Bandwidth, p.Price)
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}
