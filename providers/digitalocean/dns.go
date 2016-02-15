// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package digitalocean

import (
	"fmt"
	"sort"

	"github.com/digitalocean/godo"
	"github.com/ryanuber/columnize"
)

func (this *doProvider) ShowDomainRecords(domain string) error {
	// Load images
	client := NewDOClient(this.token)
	records, err := DomainRecordList(client, domain)
	if err != nil {
		return err
	}

	lines := []string{
		"Type | Name | Data | Priority | Weight | Port",
	}
	for _, r := range records {
		line := fmt.Sprintf("%s | %s | %s | %v | %v | %v", r.Type, trimLength(r.Name, 20), trimLength(r.Data, 60), r.Priority, r.Weight, r.Port)
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

func (this *doProvider) CreateDnsRecord(domain, _type, name, data string) error {
	client := NewDOClient(this.token)
	record := &godo.DomainRecordEditRequest{
		Type: _type,
		Name: name,
		Data: data,
	}
	_, _, err := client.Domains.CreateRecord(domain, record)
	if err != nil {
		return err
	}
	return nil
}

func (this *doProvider) DeleteDnsRecord(domain, _type, name, data string) error {
	client := NewDOClient(this.token)
	records, err := DomainRecordList(client, domain)
	if err != nil {
		return err
	}
	for _, r := range records {
		if r.Type != _type || r.Name != name {
			continue
		}
		if data != "" && r.Data != data {
			continue
		}
		// Found matching record
		_, err := client.Domains.DeleteRecord(domain, r.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
