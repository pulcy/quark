package digitalocean

import (
	"errors"
	"strings"

	"arvika.pulcy.com/iggi/droplets/providers"
)

func (this *doProvider) DeleteCluster(info *providers.ClusterInfo) error {
	droplets, err := this.getInstances(info)
	if err != nil {
		return err
	}
	client := NewDOClient(this.token)
	for _, d := range droplets {
		if strings.Contains(d.Name, "arvika") {
			return errors.New("Not allowed to delete arvika")
		}

		// Delete DNS record
		if err := this.deleteDnsRecord(info.Domain, "A", d.Name, ""); err != nil {
			return err
		}

		// Delete droplet
		_, err := client.Droplets.Delete(d.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
