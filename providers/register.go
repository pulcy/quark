package providers

import (
	"strings"

	"github.com/op/go-logging"
)

// RegisterInstance creates DNS records for an instance
func RegisterInstance(logger *logging.Logger, dnsProvider DnsProvider, options CreateInstanceOptions, name string, publicIpv4, publicIpv6 string) error {
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
