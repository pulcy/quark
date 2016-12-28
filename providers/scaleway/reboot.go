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

package scaleway

import (
	"github.com/pulcy/quark/providers"
)

// Perform a reboot of the given instance
func (vp *scalewayProvider) RebootInstance(instance providers.ClusterInstance) error {
	s, err := instance.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()
	if err := s.Sync(vp.Logger); err != nil {
		return maskAny(err)
	}
	if err := vp.client.PostServerAction(instance.ID, "reboot"); err != nil {
		vp.Logger.Errorf("reboot failed: %#v", err)
		return maskAny(err)
	}
	return nil
}
