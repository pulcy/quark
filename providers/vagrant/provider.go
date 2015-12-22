package vagrant

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/op/go-logging"

	"arvika.pulcy.com/pulcy/droplets/providers"
	"arvika.pulcy.com/pulcy/droplets/templates"
)

const (
	fileMode            = os.FileMode(0775)
	cloudConfigTemplate = "templates/cloud-config.tmpl"
	vagrantFileTemplate = "templates/Vagrantfile.tmpl"
	vagrantFileName     = "Vagrantfile"
	configTemplate      = "templates/config.rb.tmpl"
	configFileName      = "config.rb"
	userDataFileName    = "user-data"
)

type vagrantProvider struct {
	Logger *logging.Logger
	folder string
}

func NewProvider(logger *logging.Logger, folder string) providers.CloudProvider {
	return &vagrantProvider{
		Logger: logger,
		folder: folder,
	}
}

func (vp *vagrantProvider) CreateAnsibleHosts(domain string, sshPort int, developersJson string) error {
	return maskAny(NotImplementedError)
}

func (vp *vagrantProvider) ShowPlans() error {
	return maskAny(NotImplementedError)
}

func (vp *vagrantProvider) ShowRegions() error {
	return maskAny(NotImplementedError)
}

func (vp *vagrantProvider) ShowImages() error {
	return maskAny(NotImplementedError)
}

func (vp *vagrantProvider) ShowKeys() error {
	return maskAny(NotImplementedError)
}

// Create a machine instance
func (vp *vagrantProvider) CreateInstance(options *providers.CreateInstanceOptions, dnsProvider providers.DnsProvider) (providers.ClusterInstance, error) {
	return providers.ClusterInstance{}, maskAny(NotImplementedError)
}

// Create an entire cluster
func (vp *vagrantProvider) CreateCluster(options *providers.CreateClusterOptions, dnsProvider providers.DnsProvider) error {
	// Ensure folder exists
	if err := os.MkdirAll(vp.folder, fileMode|os.ModeDir); err != nil {
		return maskAny(err)
	}

	vopts := struct {
		InstanceCount int
	}{
		InstanceCount: options.InstanceCount,
	}

	// Vagrantfile
	content, err := templates.Render(vagrantFileTemplate, vopts)
	if err != nil {
		return maskAny(err)
	}
	if err := ioutil.WriteFile(filepath.Join(vp.folder, vagrantFileName), []byte(content), fileMode); err != nil {
		return maskAny(err)
	}

	// config.rb
	content, err = templates.Render(configTemplate, vopts)
	if err != nil {
		return maskAny(err)
	}
	if err := ioutil.WriteFile(filepath.Join(vp.folder, configFileName), []byte(content), fileMode); err != nil {
		return maskAny(err)
	}

	// user-data
	discoveryUrl, err := providers.NewDiscoveryUrl(options.InstanceCount)
	if err != nil {
		return maskAny(err)
	}
	instanceOptions := options.NewCreateInstanceOptions(discoveryUrl)
	opts := instanceOptions.NewCloudConfigOptions()
	opts.PrivateIPv4 = "$private_ipv4"
	opts.IncludeSshKeys = true
	opts.PrivateClusterDevice = "eth0"

	content, err = templates.Render(cloudConfigTemplate, opts)
	if err != nil {
		return maskAny(err)
	}
	if err := ioutil.WriteFile(filepath.Join(vp.folder, userDataFileName), []byte(content), fileMode); err != nil {
		return maskAny(err)
	}

	return nil
}

// Get names of instances of a cluster
func (vp *vagrantProvider) GetInstances(info *providers.ClusterInfo) ([]providers.ClusterInstance, error) {
	return nil, nil
}

// Remove all instances of a cluster
func (vp *vagrantProvider) DeleteCluster(info *providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	return maskAny(NotImplementedError)
}

func (vp *vagrantProvider) DeleteInstance(info *providers.ClusterInstanceInfo, dnsProvider providers.DnsProvider) error {
	return maskAny(NotImplementedError)
}

func (vp *vagrantProvider) ShowDomainRecords(domain string) error {
	return maskAny(NotImplementedError)
}
