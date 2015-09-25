package cloudflare

import (
	"fmt"
	"sort"

	"github.com/juju/errgo"
	"github.com/ryanuber/columnize"
)

type CfZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p *cfProvider) zones(domain string) ([]CfZone, error) {
	url := apiUrl + "zones?name=" + domain
	res, err := p.get(url, "application/json")
	if err != nil {
		return nil, maskAny(err)
	}

	zones := []CfZone{}
	if err := res.UnmarshalResult(&zones); err != nil {
		return nil, maskAny(err)
	}

	return zones, nil
}

func (p *cfProvider) zoneID(domain string) (string, error) {
	zones, err := p.zones(domain)
	if err != nil {
		return "", maskAny(err)
	}
	for _, z := range zones {
		if z.Name == domain {
			return z.ID, nil
		}
	}
	return "", maskAny(errgo.WithCausef(nil, DomainNotFoundError, domain))
}

type CfDnsRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl'`
}

func (p *cfProvider) ShowDomainRecords(domain string) error {
	id, err := p.zoneID(domain)
	if err != nil {
		return maskAny(err)
	}

	url := apiUrl + fmt.Sprintf("zones/%s/dns_records", id)
	res, err := p.get(url, "application/json")
	if err != nil {
		return maskAny(err)
	}

	records := []CfDnsRecord{}
	if err := res.UnmarshalResult(&records); err != nil {
		return maskAny(err)
	}

	lines := []string{
		"Type | Name | Data",
	}
	for _, r := range records {
		line := fmt.Sprintf("%s | %s | %s", r.Type, trimLength(r.Name, 20), trimLength(r.Content, 60))
		lines = append(lines, line)
	}

	sort.Strings(lines[1:])
	result := columnize.SimpleFormat(lines)
	fmt.Println(result)

	return nil
}

func trimLength(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	} else {
		return s
	}
}

func (p *cfProvider) CreateDnsRecord(domain, recordTpe, name, data string) error {
	return nil
}

func (p *cfProvider) DeleteDnsRecord(domain, recordType, name, data string) error {
	return nil
}
