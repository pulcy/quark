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

package templates

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"

	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

func Render(templateName string, options interface{}) (string, error) {
	asset, err := Asset(templateName)
	if err != nil {
		return "", maskAny(err)
	}

	// parse template
	var tmpl *template.Template
	tmpl = template.New(templateName)
	funcMap := template.FuncMap{
		"escape":     escape,
		"quote":      strconv.Quote,
		"yamlPrefix": yamlPrefix,
	}
	tmpl.Funcs(funcMap)
	_, err = tmpl.Parse(string(asset))
	if err != nil {
		return "", maskAny(err)
	}
	// write file to buffer
	tmpl.Funcs(funcMap)
	buffer := &bytes.Buffer{}
	err = tmpl.Execute(buffer, options)
	if err != nil {
		return "", maskAny(err)
	}

	return buffer.String(), nil
}

func escape(s string) string {
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
}

func yamlPrefix(s string, spaces int) string {
	lines := strings.Split(s, "\n")
	result := ""
	prefix := ""
	for spaces > 0 {
		prefix = prefix + " "
		spaces--
	}
	for i, l := range lines {
		if i > 0 {
			result = result + "\n" + prefix
		}
		result = result + l
	}
	return result
}
