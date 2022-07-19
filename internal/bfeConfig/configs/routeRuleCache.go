// Copyright (c) 2021 The BFE Authors.
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

package configs

import (
	"sort"

	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	netv1 "k8s.io/api/networking/v1"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/cache"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
)

type httpRule struct {
	Cluster string
	*cache.BaseRule
}

type RouteRuleCache struct {
	*cache.BaseCache
}

func NewRouteRuleCache() *RouteRuleCache {
	return &RouteRuleCache{
		BaseCache: cache.NewBaseCache(),
	}
}

func (routeRuleCache *RouteRuleCache) GetRouteRules() (basicRuleList []*httpRule, advancedRuleList []*httpRule) {
	c := routeRuleCache.BaseRules
	for _, paths := range c.RuleMap {
		for _, rules := range paths {
			if len(rules) == 0 {
				continue
			}

			// add host+path rule to basic rule list
			if len(rules) == 1 && annotations.Priority(rules[0].GetAnnotations()) == annotations.PriorityBasic {
				basicRuleList = append(basicRuleList, rules[0].(*httpRule))
				continue
			}
			// add a fake basicRule,cluster=ADVANCED_MODE
			newRule := *rules[0].(*httpRule)
			newRule.Cluster = route_rule_conf.AdvancedMode
			basicRuleList = append(basicRuleList, &newRule)

			// add advanced rule
			for _, rule := range rules {
				advancedRuleList = append(advancedRuleList, rule.(*httpRule))
			}
		}
	}

	// host: exact match over wildcard match
	// path: long path over short path
	sort.SliceStable(advancedRuleList, func(i, j int) bool {
		return advancedRuleList[i].Compare(advancedRuleList[j])
	})

	return
}

func (routeRuleCache *RouteRuleCache) UpdateByIngress(ingress *netv1.Ingress) error {
	return routeRuleCache.BaseCache.UpdateByIngressFramework(
		ingress,
		func() (bool, error) {
			return true, nil
		},
		func(ingress *netv1.Ingress, host, path string) cache.Rule {
			return &httpRule{
				BaseRule: cache.NewBaseRule(
					util.NamespacedName(ingress.Namespace, ingress.Name),
					host,
					path,
					ingress.Annotations,
					ingress.CreationTimestamp.Time,
				),
				Cluster: ingress.ClusterName,
			}
		},
		func() error {
			return nil
		},
	)
}
