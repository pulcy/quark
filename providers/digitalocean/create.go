package digitalocean

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dchest/uniuri"
	"github.com/digitalocean/godo"

	"arvika.pulcy.com/pulcy/droplets/providers"
	"arvika.pulcy.com/pulcy/droplets/templates"
)

const (
	cloudConfigTemplate = "templates/cloud-config.tmpl"
)

func (dp *doProvider) CreateCluster(options *providers.CreateClusterOptions, dnsProvider providers.DnsProvider) error {
	discoveryURL, err := providers.NewDiscoveryUrl(options.InstanceCount)
	if err != nil {
		return maskAny(err)
	}

	wg := sync.WaitGroup{}
	errors := make(chan error, options.InstanceCount)
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
				DiscoveryUrl:         discoveryURL,
				SSHKeyNames:          options.SSHKeyNames,
				YardImage:            options.YardImage,
				YardPassphrase:       options.YardPassphrase,
				StunnelPemPassphrase: options.StunnelPemPassphrase,
				FlannelNetworkCidr:   options.FlannelNetworkCidr,
				RebootStrategy:       options.RebootStrategy,
			}
			err := dp.CreateInstance(instanceOptions, dnsProvider)
			if err != nil {
				errors <- maskAny(err)
			}
		}(i)
	}
	wg.Wait()
	close(errors)
	err = <-errors
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (dp *doProvider) CreateInstance(options *providers.CreateInstanceOptions, dnsProvider providers.DnsProvider) error {
	client := NewDOClient(dp.token)

	keys := []godo.DropletCreateSSHKey{}
	listedKeys, err := KeyList(client)
	if err != nil {
		return maskAny(err)
	}
	for _, key := range options.SSHKeyNames {
		k := findKeyID(key, listedKeys)
		if k == nil {
			return maskAny(errors.New("Key not found"))
		}
		keys = append(keys, godo.DropletCreateSSHKey{ID: k.ID})
	}

	opts := providers.CloudConfigOptions{
		DiscoveryUrl:         options.DiscoveryUrl,
		Region:               options.Region,
		PrivateIPv4:          "$private_ipv4",
		YardPassphrase:       options.YardPassphrase,
		StunnelPemPassphrase: options.StunnelPemPassphrase,
		YardImage:            options.YardImage,
		FlannelNetworkCidr:   options.FlannelNetworkCidr,
		RebootStrategy:       options.RebootStrategy,
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
	dp.Logger.Info("Creating droplet: %s, %s, %s", request.Region, request.Size, options.Image)
	dp.Logger.Debug(cloudConfig)
	createDroplet, _, err := client.Droplets.Create(request)
	if err != nil {
		return maskAny(err)
	}

	// Wait for active
	dp.Logger.Info("Waiting for droplet '%s'", createDroplet.Name)
	droplet, err := dp.waitUntilDropletActive(createDroplet.ID)
	if err != nil {
		return maskAny(err)
	}

	publicIpv4 := getIpv4(*droplet, "public")
	publicIpv6 := getIpv6(*droplet, "public")
	if err := providers.RegisterInstance(dp.Logger, dnsProvider, options, createDroplet.Name, publicIpv4, publicIpv6); err != nil {
		return maskAny(err)
	}

	dp.Logger.Info("Droplet '%s' is ready", createDroplet.Name)

	return nil
}

func (dp *doProvider) waitUntilDropletActive(id int) (*godo.Droplet, error) {
	client := NewDOClient(dp.token)
	for {
		droplet, _, err := client.Droplets.Get(id)
		if err != nil {
			return nil, err
		}
		if droplet.Status != "new" {
			return droplet, nil
		}
		// Wait a while
		time.Sleep(time.Second * 5)
	}
}

func findKeyID(key string, listedKeys []godo.Key) *godo.Key {
	for _, k := range listedKeys {
		if k.Name == key {
			return &k
		}
	}
	return nil
}
