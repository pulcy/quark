package digitalocean

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ryanuber/columnize"
)

func (this *doProvider) ShowImages() error {
	// Load images
	client := NewDOClient(this.token)
	images, err := ImageList(client)
	if err != nil {
		return err
	}

	lines := []string{
		"Slug | Name | Distribution | Regions",
	}
	for _, r := range images {
		if !r.Public {
			continue
		}
		line := fmt.Sprintf("%s | %s | %s | %s", r.Slug, r.Name, r.Distribution, strings.Join(r.Regions, " "))
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}
