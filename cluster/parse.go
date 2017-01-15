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
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/juju/errgo"
)

// ParseClusterFromFile reads a cluster from file
func ParseClusterFromFile(path string) (*Cluster, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, maskAny(err)
	}
	// Parse the input
	root, err := hcl.Parse(string(data))
	if err != nil {
		return nil, maskAny(err)
	}
	// Top-level item should be a list
	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return nil, errgo.New("error parsing: root should be an object")
	}
	matches := list.Filter("cluster")
	if len(matches.Items) == 0 {
		return nil, errgo.New("'cluster' stanza not found")
	}

	// Parse hcl into Cluster
	cluster := &Cluster{}
	if err := cluster.parse(matches); err != nil {
		return nil, maskAny(err)
	}
	cluster.setDefaults()

	// Validate the cluster
	if err := cluster.validate(); err != nil {
		return nil, maskAny(err)
	}

	return cluster, nil
}

// Parse a Cluster
func (c *Cluster) parse(list *ast.ObjectList) error {
	list = list.Children()
	if len(list.Items) != 1 {
		return fmt.Errorf("only one 'cluster' block allowed")
	}

	// Get our cluster object
	obj := list.Items[0]

	// Decode the object
	excludeList := []string{
		"default-options",
		"docker",
		"orchestrator",
		"fleet",
		"kubernetes",
		"quark",
	}
	if err := Decode(obj.Val, excludeList, nil, c); err != nil {
		return maskAny(err)
	}
	c.Stack = obj.Keys[0].Token.Value().(string)

	// Value should be an object
	var listVal *ast.ObjectList
	if ot, ok := obj.Val.(*ast.ObjectType); ok {
		listVal = ot.List
	} else {
		return errgo.Newf("cluster '%s' value: should be an object", c.Stack)
	}

	// If we have quark object, parse it
	if o := listVal.Filter("quark"); len(o.Items) > 0 {
		for _, o := range o.Elem().Items {
			if obj, ok := o.Val.(*ast.ObjectType); ok {
				if err := c.QuarkOptions.parse(obj, *c); err != nil {
					return maskAny(err)
				}
			} else {
				return maskAny(errgo.WithCausef(nil, ValidationError, "docker of cluster '%s' is not an object", c.Stack))
			}
		}
	}

	return nil
}

// parse a QuarkOptions
func (options *QuarkOptions) parse(obj *ast.ObjectType, c Cluster) error {
	// Parse the object
	excludeList := []string{
		"profile",
	}
	values, err := decodeIntoMap(obj, excludeList, nil)
	if err != nil {
		return maskAny(err)
	}
	options.DefaultValues = values

	// Parse profiles
	if o := obj.List.Filter("profile"); len(o.Items) > 0 {
		for _, o := range o.Children().Items {
			if obj, ok := o.Val.(*ast.ObjectType); ok {
				p := Profile{}
				n := o.Keys[0].Token.Value().(string)
				if err := p.parse(obj); err != nil {
					return maskAny(err)
				}
				p.Name = n
				options.Profiles = append(options.Profiles, p)
			} else {
				return maskAny(errgo.WithCausef(nil, ValidationError, "profile is not an object"))
			}
		}
	}

	return nil
}

// parse a Profile
func (options *Profile) parse(obj *ast.ObjectType) error {
	// Parse the object
	excludeList := []string{}
	values, err := decodeIntoMap(obj, excludeList, nil)
	if err != nil {
		return maskAny(err)
	}
	options.Values = values

	return nil
}
