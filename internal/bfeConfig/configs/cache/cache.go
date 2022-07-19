// Copyright (c) 2022 The BFE Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import netv1 "k8s.io/api/networking/v1"

// Cache is an abstraction of the cache of Rules.
type Cache interface {
	// GetRules gets the rules stored in the cache.
	// These rules are usually ordered by priority.
	GetRules() []Rule

	// UpdateByIngress uses an ingress to update the cache
	UpdateByIngress(ingress *netv1.Ingress) error

	// DeleteRulesByIngress delete everything related to the ingress from the cache
	DeleteRulesByIngress(ingress string)

	// ContainsIngress returns true if ingress exist in cache
	ContainsIngress(ingress string) bool
}
