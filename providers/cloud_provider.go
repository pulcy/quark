package providers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dchest/uniuri"
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
	ShowRegions() error
	ShowImages() error
	ShowKeys() error
	ShowInstanceTypes() error

	// Apply defaults for the given options
	InstanceDefaults(options CreateInstanceOptions) CreateInstanceOptions

	// Apply defaults for the given options
	ClusterDefaults(options CreateClusterOptions) CreateClusterOptions

	// Create a machine instance
	CreateInstance(options CreateInstanceOptions, dnsProvider DnsProvider) (ClusterInstance, error)

	// Create an entire cluster
	CreateCluster(options CreateClusterOptions, dnsProvider DnsProvider) error

	// Get names of instances of a cluster
	GetInstances(info ClusterInfo) ([]ClusterInstance, error)

	// Remove all instances of a cluster
	DeleteCluster(info ClusterInfo, dnsProvider DnsProvider) error

	// Remove a single instance of a cluster
	DeleteInstance(info ClusterInstanceInfo, dnsProvider DnsProvider) error

	ShowDomainRecords(domain string) error
}

// ClusterInfo describes a cluster
type ClusterInfo struct {
	ID     string // /etc/pulcy/cluster-id, used for vault-monkey authentication
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

type InstanceConfig struct {
	ImageID  string // ID of the image to install on each instance
	RegionID string // ID of the region to run all instances in
	TypeID   string // ID of the type of each instance
}

func (ic InstanceConfig) String() string {
	return fmt.Sprintf("type: %s, image: %s, region: %s", ic.TypeID, ic.ImageID, ic.RegionID)
}

// Options for creating a cluster
type CreateClusterOptions struct {
	ClusterInfo
	InstanceConfig
	SSHKeyNames             []string // List of names of SSH keys to install on each instance
	SSHKeyGithubAccount     string   // Github account name used to fetch SSH keys
	InstanceCount           int      // Number of instances to start
	GluonImage              string   // Docker image containing gluon
	RebootStrategy          string
	PrivateRegistryUrl      string // URL of private docker registry
	PrivateRegistryUserName string // Username of private docker registry
	PrivateRegistryPassword string // Password of private docker registry
}

// NewCreateInstanceOptions creates a new CreateInstanceOptions instances with all
// values inherited from the given CreateClusterOptions
func (o *CreateClusterOptions) NewCreateInstanceOptions() CreateInstanceOptions {
	io := CreateInstanceOptions{
		ClusterInfo:             o.ClusterInfo,
		InstanceConfig:          o.InstanceConfig,
		SSHKeyNames:             o.SSHKeyNames,
		SSHKeyGithubAccount:     o.SSHKeyGithubAccount,
		GluonImage:              o.GluonImage,
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
	InstanceConfig
	ClusterName             string   // Full name of the cluster e.g. "dev1.example.com"
	InstanceName            string   // Name of the instance e.g. "abc123.dev1.example.com"
	SSHKeyNames             []string // List of names of SSH keys to install
	SSHKeyGithubAccount     string   // Github account name used to fetch SSH keys
	GluonImage              string   // Docker image containing gluon
	RebootStrategy          string
	PrivateRegistryUrl      string // URL of private docker registry
	PrivateRegistryUserName string // Username of private docker registry
	PrivateRegistryPassword string // Password of private docker registry
	EtcdProxy               bool   // If set, this instance will be an ETCD proxy
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
		ClusterID:               o.ClusterInfo.ID,
		FleetMetadata:           o.fleetMetadata(),
		GluonImage:              o.GluonImage,
		RebootStrategy:          o.RebootStrategy,
		PrivateRegistryUrl:      o.PrivateRegistryUrl,
		PrivateRegistryUserName: o.PrivateRegistryUserName,
		PrivateRegistryPassword: o.PrivateRegistryPassword,
	}
	return cco
}

// fleetMetadata creates a valid fleet metadata string for use in cloud-config
func (o *CreateInstanceOptions) fleetMetadata() string {
	list := []string{fmt.Sprintf("region=%s", o.RegionID)}
	return strings.Join(list, ",")
}

// Options for cloud-config files
type CloudConfigOptions struct {
	ClusterID               string
	FleetMetadata           string
	PrivateIPv4             string
	GluonImage              string
	SshKeys                 []string
	RebootStrategy          string
	PrivateClusterDevice    string
	PrivateRegistryUrl      string // URL of private docker registry
	PrivateRegistryUserName string // Username of private docker registry
	PrivateRegistryPassword string // Password of private docker registry
}

// Validate the given options
func (ic InstanceConfig) Validate() error {
	if ic.ImageID == "" {
		return errors.New("Please specific an image")
	}
	if ic.RegionID == "" {
		return errors.New("Please specific a region")
	}
	if ic.TypeID == "" {
		return errors.New("Please specific a type")
	}
	return nil
}

// Validate the given options
func (cco CreateClusterOptions) Validate() error {
	if cco.Domain == "" {
		return errors.New("Please specify a domain")
	}
	if cco.Name == "" {
		return errors.New("Please specify a name")
	}
	if strings.ContainsAny(cco.Name, ".") {
		return errors.New("Invalid characters in name")
	}
	if err := cco.InstanceConfig.Validate(); err != nil {
		return maskAny(err)
	}
	if len(cco.SSHKeyNames) == 0 {
		return errors.New("Please specify at least one SSH key")
	}
	if cco.SSHKeyGithubAccount == "" {
		return errors.New("Please specify a valid ssh key github account")
	}
	if cco.InstanceCount < 1 {
		return errors.New("Please specify a valid instance count")
	}
	if cco.GluonImage == "" {
		return errors.New("Please specify a gluon-image")
	}
	if cco.PrivateRegistryUrl == "" {
		return errors.New("Please specify a private-registry-url")
	}
	if cco.PrivateRegistryUserName == "" {
		return errors.New("Please specify a private-registry-username")
	}
	if cco.PrivateRegistryPassword == "" {
		return errors.New("Please specify a private-registry-password")
	}
	return nil
}

// Validate the given options
func (cio CreateInstanceOptions) Validate() error {
	if cio.ClusterName == "" {
		return errors.New("Please specify a cluster-name")
	}
	if cio.InstanceName == "" {
		return errors.New("Please specify a instance-name")
	}
	if err := cio.InstanceConfig.Validate(); err != nil {
		return maskAny(err)
	}
	if len(cio.SSHKeyNames) == 0 {
		return errors.New("Please specific at least one SSH key")
	}
	if cio.SSHKeyGithubAccount == "" {
		return errors.New("Please specify a valid ssh key github account")
	}
	if cio.GluonImage == "" {
		return errors.New("Please specify a gluon-image")
	}
	return nil
}
