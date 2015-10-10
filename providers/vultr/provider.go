package vultr

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JamesClonk/vultr/lib"
	"github.com/dchest/uniuri"
	"github.com/juju/errgo"
	"github.com/op/go-logging"
	"github.com/ryanuber/columnize"

	"arvika.pulcy.com/pulcy/droplets/providers"
	"arvika.pulcy.com/pulcy/droplets/templates"
)

const (
	fileMode            = os.FileMode(0775)
	cloudConfigTemplate = "templates/cloud-config.tmpl"
	bootstrapTemplate   = "templates/bootstrap.sh.tmpl"

	regionAmsterdam = 7
	coreosStable    = 179
	plan768MB       = 29
	plan1GB         = 93
)

type vultrProvider struct {
	Logger *logging.Logger
	client *lib.Client
}

func NewProvider(logger *logging.Logger, apiKey string) providers.CloudProvider {
	client := lib.NewClient(apiKey, nil)
	return &vultrProvider{
		Logger: logger,
		client: client,
	}
}

func (vp *vultrProvider) CreateAnsibleHosts(domain string, sshPort int, developersJson string) error {
	return maskAny(NotImplementedError)
}

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

func (vp *vultrProvider) ShowPlans() error {
	plans, err := vp.client.GetPlans()
	if err != nil {
		return maskAny(err)
	}

	lines := []string{
		"ID | Name | VCpu | RAM | Disk | Bandwidth | Price",
	}
	for _, p := range plans {
		line := fmt.Sprintf("%02d | %s | %d | %s | %s | %s | %s", p.ID, p.Name, p.VCpus, p.RAM, p.Disk, p.Bandwidth, p.Price)
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}

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

func (vp *vultrProvider) ShowKeys() error {
	keys, err := vp.client.GetSSHKeys()
	if err != nil {
		return maskAny(err)
	}
	lines := []string{
		"ID | Name | Public-key",
	}
	for _, r := range keys {
		line := fmt.Sprintf("%v | %s | %s", r.ID, r.Name, r.Key)
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}

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
		StunnelPemPassphrase: options.StunnelPemPassphrase,
		YardImage:            options.YardImage,
		FlannelNetworkCidr:   options.FlannelNetworkCidr,
		IncludeSshKeys:       true,
		RebootStrategy:       options.RebootStrategy,
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
	discoveryUrl, err := providers.NewDiscoveryUrl(options.InstanceCount)
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
				DiscoveryUrl:         discoveryUrl,
				SSHKeyNames:          options.SSHKeyNames,
				YardImage:            options.YardImage,
				YardPassphrase:       options.YardPassphrase,
				StunnelPemPassphrase: options.StunnelPemPassphrase,
				FlannelNetworkCidr:   options.FlannelNetworkCidr,
				RebootStrategy:       options.RebootStrategy,
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

// Search for an SSH key with given name and return its ID
func (vp *vultrProvider) findSSHKeyID(keyName string) (string, error) {
	keys, err := vp.client.GetSSHKeys()
	if err != nil {
		return "", maskAny(err)
	}
	for _, k := range keys {
		if k.Name == keyName {
			return k.ID, nil
		}
	}
	return "", errgo.WithCausef(nil, InvalidArgumentError, "key %s not found", keyName)
}

// createBootstrap creates a bootstrap file
func (vp *vultrProvider) createBootstrap(options *providers.CreateInstanceOptions) (string, error) {
	// user-data
	ccOpts := providers.CloudConfigOptions{
		DiscoveryUrl:         options.DiscoveryUrl,
		Region:               options.Region,
		PrivateIPv4:          "$private_ipv4",
		YardPassphrase:       options.YardPassphrase,
		StunnelPemPassphrase: options.StunnelPemPassphrase,
		YardImage:            options.YardImage,
		FlannelNetworkCidr:   options.FlannelNetworkCidr,
		IncludeSshKeys:       true,
		RebootStrategy:       options.RebootStrategy,
	}
	content, err := templates.Render(cloudConfigTemplate, ccOpts)
	if err != nil {
		return "", maskAny(err)
	}

	// bootstrap.sh
	bsOpts := struct {
		CloudConfig string
	}{
		CloudConfig: content,
	}
	bootstrap, err := templates.Render(bootstrapTemplate, bsOpts)
	if err != nil {
		return "", maskAny(err)
	}
	bootstrapFile, err := ioutil.TempFile("", "bootstrap")
	if err != nil {
		return "", maskAny(err)
	}
	defer bootstrapFile.Close()
	if _, err := bootstrapFile.WriteString(bootstrap); err != nil {
		return "", maskAny(err)
	}

	return bootstrapFile.Name(), nil
}

// Get names of instances of a cluster
func (vp *vultrProvider) GetInstances(info *providers.ClusterInfo) ([]providers.ClusterInstance, error) {
	servers, err := vp.getInstances(info)
	if err != nil {
		return nil, maskAny(err)
	}
	list := []providers.ClusterInstance{}
	for _, s := range servers {
		info := providers.ClusterInstance{
			Name:        s.Name,
			PrivateIpv4: s.InternalIP,
			PublicIpv4:  s.MainIP,
		}
		list = append(list, info)

	}
	return list, nil
}

func (vp *vultrProvider) getInstances(info *providers.ClusterInfo) ([]lib.Server, error) {
	servers, err := vp.client.GetServers()
	if err != nil {
		return nil, maskAny(err)
	}

	postfix := fmt.Sprintf(".%s.%s", info.Name, info.Domain)
	result := []lib.Server{}
	for _, s := range servers {
		if strings.HasSuffix(s.Name, postfix) {
			result = append(result, s)
		}
	}

	return result, nil
}

// Remove all instances of a cluster
func (vp *vultrProvider) DeleteCluster(info *providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	return maskAny(NotImplementedError)
}

func (vp *vultrProvider) ShowDomainRecords(domain string) error {
	return maskAny(NotImplementedError)
}
