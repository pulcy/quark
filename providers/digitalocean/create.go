package digitalocean

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dchest/uniuri"
	"github.com/digitalocean/godo"
	"github.com/juju/errgo"

	"arvika.pulcy.com/iggi/droplets/providers"
	"arvika.pulcy.com/iggi/droplets/templates"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

const (
	cloudConfigTemplate = "templates/cloud-config.tmpl"
)

func (this *doProvider) CreateCluster(options *providers.CreateClusterOptions, dnsProvider providers.DnsProvider) error {
	discoveryUrl, err := newDiscoveryUrl()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	errors := make(chan error, options.InstanceCount)
	defer close(errors)
	for i := 1; i <= options.InstanceCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			prefix := strings.ToLower(uniuri.NewLen(8))
			instanceOptions := &providers.CreateInstanceOptions{
				Domain:               options.Domain,
				ClusterName:          fmt.Sprintf("%s.%s", options.Name, options.Domain),
				InstanceName:         fmt.Sprintf("%s.%s.%s", prefix, options.Name, options.Domain),
				Region:               options.Region,
				Image:                options.Image,
				Size:                 options.Size,
				DiscoveryUrl:         discoveryUrl,
				SSHKeyNames:          options.SSHKeyNames,
				YardImage:            options.YardImage,
				YardPassphrase:       options.YardPassphrase,
				StunnelPemPassphrase: options.StunnelPemPassphrase,
				WeavePassword:        options.WeavePassword,
			}
			err := this.CreateInstance(instanceOptions, dnsProvider)
			if err != nil {
				errors <- err
			}
		}(i)
	}
	wg.Wait()
	err = <-errors
	if err != nil {
		return err
	}

	return nil
}

func (this *doProvider) CreateInstance(options *providers.CreateInstanceOptions, dnsProvider providers.DnsProvider) error {
	client := NewDOClient(this.token)

	keys := []godo.DropletCreateSSHKey{}
	listedKeys, err := KeyList(client)
	if err != nil {
		return err
	}
	for _, key := range options.SSHKeyNames {
		k := findKeyId(key, listedKeys)
		if k == nil {
			return errors.New("Key not found")
		}
		keys = append(keys, godo.DropletCreateSSHKey{ID: k.ID})
	}

	opts := struct {
		DiscoveryUrl         string
		Region               string
		PrivateIPv4          string
		YardPassphrase       string
		StunnelPemPassphrase string
		YardImage            string
		WeavePassword        string
	}{
		DiscoveryUrl:         options.DiscoveryUrl,
		Region:               options.Region,
		PrivateIPv4:          "$private_ipv4",
		YardPassphrase:       options.YardPassphrase,
		StunnelPemPassphrase: options.StunnelPemPassphrase,
		YardImage:            options.YardImage,
		WeavePassword:        options.WeavePassword,
	}
	cloudConfig, err := templates.Render(cloudConfigTemplate, opts)
	if err != nil {
		return maskAny(err)
	}

	request := &godo.DropletCreateRequest{
		Name:              options.InstanceName,
		Region:            options.Region,
		Size:              options.Size,
		Image:             godo.DropletCreateImage{Slug: options.Image},
		SSHKeys:           keys,
		Backups:           false,
		IPv6:              true,
		PrivateNetworking: true,
		UserData:          cloudConfig,
	}

	// Create droplet
	createDroplet, _, err := client.Droplets.Create(request)
	if err != nil {
		return err
	}

	// Wait for active
	droplet, err := this.waitUntilDropletActive(createDroplet.ID)
	if err != nil {
		return nil
	}

	publicIpv4 := getIpv4(*droplet, "public")
	publicIpv6 := getIpv6(*droplet, "public")
	fmt.Printf("%s: %s: %s\n", droplet.Name, publicIpv4, publicIpv6)

	// Create DNS record for the instance
	if err := dnsProvider.CreateDnsRecord(options.Domain, "A", options.InstanceName, publicIpv4); err != nil {
		return err
	}
	if err := dnsProvider.CreateDnsRecord(options.Domain, "A", options.ClusterName, publicIpv4); err != nil {
		return err
	}
	if publicIpv6 != "" {
		if err := dnsProvider.CreateDnsRecord(options.Domain, "AAAA", options.InstanceName, publicIpv6); err != nil {
			return err
		}
		if err := dnsProvider.CreateDnsRecord(options.Domain, "AAAA", options.ClusterName, publicIpv6); err != nil {
			return err
		}
	}

	return nil
}

func (this *doProvider) waitUntilDropletActive(id int) (*godo.Droplet, error) {
	client := NewDOClient(this.token)
	for {
		droplet, _, err := client.Droplets.Get(id)
		if err != nil {
			return nil, err
		}
		if droplet.Status != "new" {
			return droplet, nil
		}
		// Wait a while
		time.Sleep(time.Millisecond * 250)
	}
}

func findKeyId(key string, listedKeys []godo.Key) *godo.Key {
	for _, k := range listedKeys {
		if k.Name == key {
			return &k
		}
	}
	return nil
}

func newDiscoveryUrl() (string, error) {
	resp, err := http.Get("https://discovery.etcd.io/new")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(body)), nil
}
