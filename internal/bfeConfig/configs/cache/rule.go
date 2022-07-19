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

import (
	"time"

	netv1 "k8s.io/api/networking/v1"
)

// Rule is an abstraction of BFE Rule.
// The BFE Rule is the basis for BFE Engine to process a Request.
// For example, the route module need to use route rules to detect which
// backend service to route a Request to;
// the redirect module need to use redirect rules to decide
// if a Request should be redirected, and how.
// The route rule and redirect rule are both Rules.
type Rule interface {
	// GetIngress gets the namespaced name of the ingress to which the Rule belongs
	GetIngress() string

	// GetHost gets the host of the Rule
	GetHost() string

	// GetPath gets the path of the Rule
	GetPath() string

	// GetAnnotations gets the annotations of the ingress to which the Rule belongs
	GetAnnotations() map[string]string

	// GetCreateTime gets the created time of the Rule
	GetCreateTime() time.Time

	// Compare will be used to prioritize rules
	Compare(anotherRule Rule) bool

	// GetCond generates a BFE Condition from the Rule
	GetCond() (string, error)
}

type BuildRuleFunc func(ingress *netv1.Ingress, host, path string) Rule
