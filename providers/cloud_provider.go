package providers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/op/go-logging"

	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

// DnsProvider holds all functions to be implemented by DNS providers
type DnsProvider interface {
	ShowDomainRecords(domain string) error
	CreateDnsRecord(domain, recordTpe, name, data string) error
	DeleteDnsRecord(domain, recordType, name, data string) error
}

// CloudProvider holds all functions to be implemented by cloud providers
type CloudProvider interface {
	CreateAnsibleHosts(domain string, sshPort int, developersJson string) error
	ShowRegions() error
	ShowImages() error
	ShowKeys() error
	ShowPlans() error

	// Create a machine instance
	CreateInstance(options *CreateInstanceOptions, dnsProvider DnsProvider) error

	// Create an entire cluster
	CreateCluster(options *CreateClusterOptions, dnsProvider DnsProvider) error

	// Get names of instances of a cluster
	GetInstances(info *ClusterInfo) ([]ClusterInstance, error)

	// Remove all instances of a cluster
	DeleteCluster(info *ClusterInfo, dnsProvider DnsProvider) error

	ShowDomainRecords(domain string) error
}

// ClusterInfo describes a cluster
type ClusterInfo struct {
	Domain string // Domain postfix (e.g. pulcy.com)
	Name   string // Name of the cluster
}

// ClusterInstance describes a single instance
type ClusterInstance struct {
	Name        string
	PrivateIpv4 string
	PublicIpv4  string
	PublicIpv6  string
}

// Options for creating a cluster
type CreateClusterOptions struct {
	ClusterInfo
	Image          string   // Name of the image to install on each instance
	Region         string   // Name of the region to run all instances in
	Size           string   // Size of each instance
	SSHKeyNames    []string // List of names of SSH keys to install on each instance
	InstanceCount  int      // Number of instances to start
	YardPassphrase string   // Passphrase for decrypting yard
	YardImage      string   // Docker image containing encrypted yard
	RebootStrategy string
}

// Options for creating an instance
type CreateInstanceOptions struct {
	Domain         string   // Name of the domain e.g. "example.com"
	ClusterName    string   // Full name of the cluster e.g. "dev1.example.com"
	InstanceName   string   // Name of the instance e.g. "abc123.dev1.example.com"
	Image          string   // Name of the image to install on the instance
	Region         string   // Name of the region to run the instance in
	Size           string   // Size of the instance
	SSHKeyNames    []string // List of names of SSH keys to install
	DiscoveryUrl   string   // Discovery url for ETCD
	YardPassphrase string   // Passphrase for decrypting yard
	YardImage      string   // Docker image containing encrypted yard
	RebootStrategy string
}

// Options for cloud-config files
type CloudConfigOptions struct {
	DiscoveryUrl         string
	Region               string
	PrivateIPv4          string
	YardPassphrase       string
	StunnelPemPassphrase string
	YardImage            string
	FlannelNetworkCidr   string
	IncludeSshKeys       bool
	RebootStrategy       string
	PrivateClusterDevice string
}

// Validate the given options
func (this *CreateClusterOptions) Validate() error {
	if this.Domain == "" {
		return errors.New("Please specific a domain")
	}
	if this.Name == "" {
		return errors.New("Please specific a name")
	}
	if strings.ContainsAny(this.Name, ".") {
		return errors.New("Invalid characters in name")
	}
	if this.Image == "" {
		return errors.New("Please specific an image")
	}
	if this.Region == "" {
		return errors.New("Please specific a region")
	}
	if this.Size == "" {
		return errors.New("Please specific a size")
	}
	if this.SSHKeyNames == nil || len(this.SSHKeyNames) == 0 {
		return errors.New("Please specific at least one SSH key")
	}
	if this.InstanceCount < 1 {
		return errors.New("Please specific a valid instance count")
	}
	if this.YardImage == "" {
		return errors.New("Please specific a yard-image")
	}
	if this.YardPassphrase == "" {
		return errors.New("Please specific a yard-passphrase")
	}
	return nil
}

// Validate the given options
func (this *CreateInstanceOptions) Validate() error {
	if this.ClusterName == "" {
		return errors.New("Please specific a cluster-name")
	}
	if this.InstanceName == "" {
		return errors.New("Please specific a instance-name")
	}
	if this.Image == "" {
		return errors.New("Please specific an image")
	}
	if this.Region == "" {
		return errors.New("Please specific a region")
	}
	if this.Size == "" {
		return errors.New("Please specific a size")
	}
	if this.SSHKeyNames == nil || len(this.SSHKeyNames) == 0 {
		return errors.New("Please specific at least one SSH key")
	}
	if this.DiscoveryUrl == "" {
		return errors.New("Please specific a discovery URL")
	}
	if this.YardImage == "" {
		return errors.New("Please specific a yard-image")
	}
	if this.YardPassphrase == "" {
		return errors.New("Please specific a yard-passphrase")
	}
	return nil
}

// NewDiscoveryUrl creates a new ETCD discovery URL
func NewDiscoveryUrl(instanceCount int) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://discovery.etcd.io/new?size=%d", instanceCount))
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

// RegisterInstance creates DNS records for an instance
func RegisterInstance(logger *logging.Logger, dnsProvider DnsProvider, options *CreateInstanceOptions, name string, publicIpv4, publicIpv6 string) error {
	logger.Info("%s: %s: %s", name, publicIpv4, publicIpv6)

	// Create DNS record for the instance
	logger.Info("Creating DNS records '%s'", name)
	if err := dnsProvider.CreateDnsRecord(options.Domain, "A", options.InstanceName, publicIpv4); err != nil {
		return maskAny(err)
	}
	if err := dnsProvider.CreateDnsRecord(options.Domain, "A", options.ClusterName, publicIpv4); err != nil {
		return maskAny(err)
	}
	if publicIpv6 != "" {
		if err := dnsProvider.CreateDnsRecord(options.Domain, "AAAA", options.InstanceName, publicIpv6); err != nil {
			return maskAny(err)
		}
		if err := dnsProvider.CreateDnsRecord(options.Domain, "AAAA", options.ClusterName, publicIpv6); err != nil {
			return maskAny(err)
		}
	}

	return nil
}

// UnRegisterInstance removes DNS records for an instance
func UnRegisterInstance(logger *logging.Logger, dnsProvider DnsProvider, instance ClusterInstance, domain string) error {
	// Delete DNS instance records
	if err := dnsProvider.DeleteDnsRecord(domain, "A", instance.Name, ""); err != nil {
		return maskAny(err)
	}
	if err := dnsProvider.DeleteDnsRecord(domain, "AAAA", instance.Name, ""); err != nil {
		return maskAny(err)
	}

	// Delete DNS cluster records
	parts := strings.Split(instance.Name, ".")
	clusterName := strings.Join(parts[1:], ".")
	if err := dnsProvider.DeleteDnsRecord(domain, "A", clusterName, instance.PublicIpv4); err != nil {
		return maskAny(err)
	}
	if err := dnsProvider.DeleteDnsRecord(domain, "AAAA", clusterName, instance.PublicIpv6); err != nil {
		return maskAny(err)
	}

	return nil
}
