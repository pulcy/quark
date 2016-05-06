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

package cluster

import (
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/juju/errgo"
	"github.com/mitchellh/mapstructure"
)

func ParseStringList(o *ast.ObjectList, context string) ([]string, error) {
	result := []string{}
	for _, o := range o.Elem().Items {
		if olit, ok := o.Val.(*ast.LiteralType); ok && olit.Token.Type == token.STRING {
			result = append(result, olit.Token.Value().(string))
		} else if list, ok := o.Val.(*ast.ListType); ok {
			for _, n := range list.List {
				if olit, ok := n.(*ast.LiteralType); ok && olit.Token.Type == token.STRING {
					result = append(result, olit.Token.Value().(string))
				} else {
					return nil, maskAny(errgo.WithCausef(nil, ValidationError, "element of %s is not a string but %v", context, n))
				}
			}
		} else {
			return nil, maskAny(errgo.WithCausef(nil, ValidationError, "%s is not a string or array", context))
		}
	}
	return result, nil
}

// decodeIntoMap from object to a map
func decodeIntoMap(obj ast.Node, excludeKeys []string, defaultValues map[string]interface{}) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := hcl.DecodeObject(&m, obj); err != nil {
		return nil, maskAny(err)
	}
	for _, key := range excludeKeys {
		delete(m, key)
	}
	for k, v := range defaultValues {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
	return m, nil
}

// Decode from object to data structure using `mapstructure`
func Decode(obj ast.Node, excludeKeys []string, defaultValues map[string]interface{}, data interface{}) error {
	var m map[string]interface{}
	if err := hcl.DecodeObject(&m, obj); err != nil {
		return maskAny(err)
	}
	for _, key := range excludeKeys {
		delete(m, key)
	}
	for k, v := range defaultValues {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
	decoderConfig := &mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		Metadata:         nil,
		Result:           data,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return maskAny(err)
	}
	return maskAny(decoder.Decode(m))
}
