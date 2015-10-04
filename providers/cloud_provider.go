package providers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type DnsProvider interface {
	ShowDomainRecords(domain string) error
	CreateDnsRecord(domain, recordTpe, name, data string) error
	DeleteDnsRecord(domain, recordType, name, data string) error
}

type CloudProvider interface {
	CreateAnsibleHosts(domain string, sshPort int, developersJson string) error
	ShowRegions() error
	ShowImages() error
	ShowKeys() error

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

type ClusterInfo struct {
	Domain string // Domain postfix (e.g. pulcy.com)
	Name   string // Name of the cluster
}

type ClusterInstance struct {
	Name        string
	PrivateIpv4 string
	PublicIpv4  string
}

type CreateClusterOptions struct {
	ClusterInfo
	Image                string   // Name of the image to install on each instance
	Region               string   // Name of the region to run all instances in
	Size                 string   // Size of each instance
	SSHKeyNames          []string // List of names of SSH keys to install on each instance
	InstanceCount        int      // Number of instances to start
	YardPassphrase       string   // Passphrase for decrypting yard
	YardImage            string   // Docker image containing encrypted yard
	StunnelPemPassphrase string   // Passphrase for decrypting stunnel.pem
	FlannelNetworkCidr   string   // CIDR used by flannel to configure the docker network bridge
}

type CreateInstanceOptions struct {
	Domain               string   // Name of the domain e.g. "example.com"
	ClusterName          string   // Full name of the cluster e.g. "dev1.example.com"
	InstanceName         string   // Name of the instance e.g. "abc123.dev1.example.com"
	Image                string   // Name of the image to install on the instance
	Region               string   // Name of the region to run the instance in
	Size                 string   // Size of the instance
	SSHKeyNames          []string // List of names of SSH keys to install
	DiscoveryUrl         string   // Discovery url for ETCD
	YardPassphrase       string   // Passphrase for decrypting yard
	YardImage            string   // Docker image containing encrypted yard
	StunnelPemPassphrase string   // Passphrase for decrypting stunnel.pem
	FlannelNetworkCidr   string   // CIDR used by flannel to configure the docker network bridge
}

type CloudConfigOptions struct {
	DiscoveryUrl         string
	Region               string
	PrivateIPv4          string
	YardPassphrase       string
	StunnelPemPassphrase string
	YardImage            string
	FlannelNetworkCidr   string
	IncludeSshKeys       bool
}

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
	if this.StunnelPemPassphrase == "" {
		return errors.New("Please specific a stunnel-pem-passphrase")
	}
	if this.FlannelNetworkCidr == "" {
		return errors.New("Please specific a flannel-network-cidr")
	}
	return nil
}

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
	if this.FlannelNetworkCidr == "" {
		return errors.New("Please specific a flannel-network-cidr")
	}
	return nil
}

func NewDiscoveryUrl() (string, error) {
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
