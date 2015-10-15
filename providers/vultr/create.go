package vultr

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/JamesClonk/vultr/lib"
	"github.com/dchest/uniuri"

	"arvika.pulcy.com/pulcy/droplets/providers"
	"arvika.pulcy.com/pulcy/droplets/templates"
)

const (
	fileMode            = os.FileMode(0775)
	cloudConfigTemplate = "templates/cloud-config.tmpl"

	regionAmsterdam = 7
	coreosStable    = 179
	plan768MB       = 29
	plan1GB         = 93
)

// Create a machine instance
func (vp *vultrProvider) CreateInstance(options *providers.CreateInstanceOptions, dnsProvider providers.DnsProvider) error {
	// Create server
	id, err := vp.createServer(options)
	if err != nil {
		return maskAny(err)
	}

	// Wait for the server to be active
	server, err := vp.waitUntilServerActive(id)
	if err != nil {
		return maskAny(err)
	}

	publicIpv4 := server.MainIP
	publicIpv6 := server.MainIPV6
	if err := providers.RegisterInstance(vp.Logger, dnsProvider, options, server.Name, publicIpv4, publicIpv6); err != nil {
		return maskAny(err)
	}

	vp.Logger.Info("Server '%s' is ready", server.Name)

	return nil
}

// Create a single server
func (vp *vultrProvider) createServer(options *providers.CreateInstanceOptions) (string, error) {
	// Find SSH key ID
	var sshid string
	if len(options.SSHKeyNames) > 0 {
		var err error
		sshid, err = vp.findSSHKeyID(options.SSHKeyNames[0])
		if err != nil {
			return "", maskAny(err)
		}
	}
	// Create cloud-config
	// user-data
	ccOpts := providers.CloudConfigOptions{
		DiscoveryUrl:         options.DiscoveryUrl,
		Region:               options.Region,
		PrivateIPv4:          "$private_ipv4",
		YardPassphrase:       options.YardPassphrase,
		YardImage:            options.YardImage,
		IncludeSshKeys:       true,
		RebootStrategy:       options.RebootStrategy,
		PrivateClusterDevice: "eth1",
	}
	userData, err := templates.Render(cloudConfigTemplate, ccOpts)
	if err != nil {
		return "", maskAny(err)
	}

	name := options.InstanceName
	opts := &lib.ServerOptions{
		IPV6:              true,
		PrivateNetworking: true,
		SSHKey:            sshid,
		UserData:          userData,
	}
	regionID := regionAmsterdam
	planID := plan768MB
	osID := coreosStable
	server, err := vp.client.CreateServer(name, regionID, planID, osID, opts)
	if err != nil {
		vp.Logger.Debug("CreateServer failed: %#v", err)
		return "", maskAny(err)
	}
	vp.Logger.Info("Server %s %s %s\n", server.ID, server.Name, server.Status)

	return server.ID, nil
}

func (vp *vultrProvider) waitUntilServerActive(id string) (lib.Server, error) {
	for {
		server, err := vp.client.GetServer(id)
		if err != nil {
			return lib.Server{}, err
		}
		if server.Status == "active" {
			return server, nil
		}
		// Wait a while
		time.Sleep(time.Second * 5)
	}
}

// Create an entire cluster
func (vp *vultrProvider) CreateCluster(options *providers.CreateClusterOptions, dnsProvider providers.DnsProvider) error {
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
				Domain:         options.Domain,
				ClusterName:    fmt.Sprintf("%s.%s", options.Name, options.Domain),
				InstanceName:   fmt.Sprintf("%s.%s.%s", prefix, options.Name, options.Domain),
				Region:         options.Region,
				Image:          options.Image,
				Size:           options.Size,
				DiscoveryUrl:   discoveryURL,
				SSHKeyNames:    options.SSHKeyNames,
				YardImage:      options.YardImage,
				YardPassphrase: options.YardPassphrase,
				RebootStrategy: options.RebootStrategy,
			}
			err := vp.CreateInstance(instanceOptions, dnsProvider)
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
