package providers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/juju/errgo"
	"github.com/op/go-logging"
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
	CreateInstance(options *CreateInstanceOptions, dnsProvider DnsProvider) (ClusterInstance, error)

	// Create an entire cluster
	CreateCluster(options *CreateClusterOptions, dnsProvider DnsProvider) error

	// Get names of instances of a cluster
	GetInstances(info *ClusterInfo) ([]ClusterInstance, error)

	// Remove all instances of a cluster
	DeleteCluster(info *ClusterInfo, dnsProvider DnsProvider) error

	// Remove a single instance of a cluster
	DeleteInstance(info *ClusterInstanceInfo, dnsProvider DnsProvider) error

	ShowDomainRecords(domain string) error
}

// ClusterInfo describes a cluster
type ClusterInfo struct {
	Domain string // Domain postfix (e.g. pulcy.com)
	Name   string // Name of the cluster
}

func (ci ClusterInfo) String() string {
	return fmt.Sprintf("%s.%s", ci.Name, ci.Domain)
}

// ClusterInstanceInfo describes a single instance of a cluster
type ClusterInstanceInfo struct {
	ClusterInfo
	Prefix string // Prefix on the instance name
}

func (cii ClusterInstanceInfo) String() string {
	return fmt.Sprintf("%s.%s.%s", cii.Prefix, cii.Name, cii.Domain)
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
	Image                   string   // Name of the image to install on each instance
	Region                  string   // Name of the region to run all instances in
	Size                    string   // Size of each instance
	SSHKeyNames             []string // List of names of SSH keys to install on each instance
	InstanceCount           int      // Number of instances to start
	YardPassphrase          string   // Passphrase for decrypting yard
	YardImage               string   // Docker image containing encrypted yard
	RebootStrategy          string
	PrivateRegistryUrl      string // URL of private docker registry
	PrivateRegistryUserName string // Username of private docker registry
	PrivateRegistryPassword string // Password of private docker registry
}

// NewCreateInstanceOptions creates a new CreateInstanceOptions instances with all
// values inherited from the given CreateClusterOptions
func (o *CreateClusterOptions) NewCreateInstanceOptions(discoveryURL string) CreateInstanceOptions {
	io := CreateInstanceOptions{
		DiscoveryURL:            discoveryURL,
		ClusterInfo:             o.ClusterInfo,
		Image:                   o.Image,
		Region:                  o.Region,
		Size:                    o.Size,
		SSHKeyNames:             o.SSHKeyNames,
		YardPassphrase:          o.YardPassphrase,
		YardImage:               o.YardImage,
		RebootStrategy:          o.RebootStrategy,
		PrivateRegistryUrl:      o.PrivateRegistryUrl,
		PrivateRegistryUserName: o.PrivateRegistryUserName,
		PrivateRegistryPassword: o.PrivateRegistryPassword,
	}
	io.SetupNames(o.Name, o.Domain)
	return io
}

// CreateInstanceOptions contains all options for creating an instance
type CreateInstanceOptions struct {
	ClusterInfo
	ClusterName             string   // Full name of the cluster e.g. "dev1.example.com"
	InstanceName            string   // Name of the instance e.g. "abc123.dev1.example.com"
	Image                   string   // Name of the image to install on the instance
	Region                  string   // Name of the region to run the instance in
	Size                    string   // Size of the instance
	SSHKeyNames             []string // List of names of SSH keys to install
	DiscoveryURL            string   // Discovery url for ETCD2
	YardPassphrase          string   // Passphrase for decrypting yard
	YardImage               string   // Docker image containing encrypted yard
	RebootStrategy          string
	PrivateRegistryUrl      string // URL of private docker registry
	PrivateRegistryUserName string // Username of private docker registry
	PrivateRegistryPassword string // Password of private docker registry
}

// SetupNames configured the ClusterName and InstanceName of the given options
// using the given cluster & domain name
func (o *CreateInstanceOptions) SetupNames(clusterName, domain string) {
	prefix := strings.ToLower(uniuri.NewLen(8))
	o.ClusterName = fmt.Sprintf("%s.%s", clusterName, domain)
	o.InstanceName = fmt.Sprintf("%s.%s.%s", prefix, clusterName, domain)
}

// NewCloudConfigOptions creates a new CloudConfigOptions instances with all
// values inherited from the given CreateInstanceOptions
func (o *CreateInstanceOptions) NewCloudConfigOptions() CloudConfigOptions {
	cco := CloudConfigOptions{
		DiscoveryURL:            o.DiscoveryURL,
		FleetMetadata:           o.fleetMetadata(),
		YardPassphrase:          o.YardPassphrase,
		YardImage:               o.YardImage,
		RebootStrategy:          o.RebootStrategy,
		PrivateRegistryUrl:      o.PrivateRegistryUrl,
		PrivateRegistryUserName: o.PrivateRegistryUserName,
		PrivateRegistryPassword: o.PrivateRegistryPassword,
	}
	return cco
}

// fleetMetadata creates a valid fleet metadata string for use in cloud-config
func (o *CreateInstanceOptions) fleetMetadata() string {
	list := []string{fmt.Sprintf("region=%s", o.Region)}
	return strings.Join(list, ",")
}

// Options for cloud-config files
type CloudConfigOptions struct {
	DiscoveryURL            string
	FleetMetadata           string
	PrivateIPv4             string
	YardPassphrase          string
	YardImage               string
	FlannelNetworkCidr      string
	IncludeSshKeys          bool
	RebootStrategy          string
	PrivateClusterDevice    string
	PrivateRegistryUrl      string // URL of private docker registry
	PrivateRegistryUserName string // Username of private docker registry
	PrivateRegistryPassword string // Password of private docker registry
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
	if this.PrivateRegistryUrl == "" {
		return errors.New("Please specific a private-registry-url")
	}
	if this.PrivateRegistryUserName == "" {
		return errors.New("Please specific a private-registry-username")
	}
	if this.PrivateRegistryPassword == "" {
		return errors.New("Please specific a private-registry-password")
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
	if this.DiscoveryURL == "" {
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
	logger.Info("Creating DNS records: '%s', '%s'", options.InstanceName, options.ClusterName)
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

// UpdateClusterMembers updates /etc/yard-cluster-members on all instances of the cluster
func UpdateClusterMembers(log *logging.Logger, info ClusterInfo, provider CloudProvider) error {
	// See if there are already instances for the given cluster
	instances, err := provider.GetInstances(&info)
	if err != nil {
		return maskAny(err)
	}

	// Update existing members
	clusterMembers := []ClusterMember{}
	for _, i := range instances {
		machineID, err := i.GetMachineID(log)
		if err != nil {
			return maskAny(err)
		}
		clusterMembers = append(clusterMembers, ClusterMember{MachineID: machineID, PrivateIP: i.PrivateIpv4})
	}
	for _, i := range instances {
		if err := i.UpdateClusterMembers(log, clusterMembers); err != nil {
			return maskAny(err)
		}
	}
	return nil
}
