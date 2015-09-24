package digitalocean

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ryanuber/columnize"
)

func (this *doProvider) ShowRegions() error {
	// Load regions
	client := NewDOClient(this.token)
	regions, err := RegionList(client)
	if err != nil {
		return err
	}

	lines := []string{
		"Slug | Name | Features | Size",
	}
	for _, r := range regions {
		if !r.Available {
			continue
		}
		line := fmt.Sprintf("%s | %s | %s | %s", r.Slug, r.Name, strings.Join(r.Features, " "), strings.Join(r.Sizes, " "))
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}
