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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jwangsadinata/go-multimap/setmultimap"
	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
)

type httpBaseCache struct {
	// ingress -> rules
	ingress2Rule *setmultimap.MultiMap

	// host -> path -> rule
	RuleMap map[string]map[string][]Rule
}

type BaseCache struct {
	BaseRules httpBaseCache
}

func NewBaseCache() *BaseCache {
	return &BaseCache{
		httpBaseCache{
			ingress2Rule: setmultimap.New(),
			RuleMap:      make(map[string]map[string][]Rule),
		},
	}
}

func NewBaseRule(ingress string, host string, path string, annots map[string]string, time time.Time) *BaseRule {
	return &BaseRule{
		Ingress:     ingress,
		Host:        host,
		Path:        path,
		Annotations: annots,
		CreateTime:  time,
	}
}

func (c *BaseCache) PutRule(rule Rule) error {
	return c.BaseRules.put(rule)
}

func (c *BaseCache) GetRules() []Rule {
	var ruleList []Rule
	for _, paths := range c.BaseRules.RuleMap {
		for _, rules := range paths {
			if len(rules) == 0 {
				continue
			}
			ruleList = append(ruleList, rules...)
		}
	}
	sort.SliceStable(ruleList, func(i, j int) bool {
		return CompareRule(ruleList[i], ruleList[j])
	})
	return ruleList
}

func (c *BaseCache) DeleteByIngress(ingress string) {
	c.BaseRules.delete(ingress)
}

// ContainsIngress returns true if ingress exist in cache
func (c *BaseCache) ContainsIngress(ingress string) bool {
	return c.BaseRules.ingress2Rule.ContainsKey(ingress)
}

func (c *BaseCache) UpdateByIngress(_ *netv1.Ingress) error {
	panic("should be implemented")
}

func (c *BaseCache) UpdateByIngressFramework(ingress *netv1.Ingress, beforeUpdate func() (bool, error), newRuleFunc BuildRuleFunc, afterUpdate func() error) error {
	if beforeUpdate != nil {
		if ok, err := beforeUpdate(); err != nil {
			return err
		} else if !ok {
			return nil
		}
	}

	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil || len(rule.HTTP.Paths) == 0 {
			continue
		}

		for _, p := range rule.HTTP.Paths {
			if err := c.addRuleToBaseCache(ingress, rule.Host, p, newRuleFunc); err != nil {
				return err
			}
		}
	}

	if afterUpdate != nil {
		return afterUpdate()
	}
	return nil
}

func (c *BaseCache) addRuleToBaseCache(ingress *netv1.Ingress, host string, httpPath netv1.HTTPIngressPath, newRuleFunc BuildRuleFunc) error {
	if err := checkHost(host); err != nil {
		return err
	}

	if len(host) == 0 {
		host = "*"
	}

	path := httpPath.Path
	if err := checkPath(path); err != nil {
		return err
	}

	if httpPath.PathType == nil || *httpPath.PathType == netv1.PathTypePrefix || *httpPath.PathType == netv1.PathTypeImplementationSpecific {
		path = path + "*"
	}
	rule := newRuleFunc(ingress, host, path)
	return c.BaseRules.put(rule)
}

func (c *httpBaseCache) delete(ingressName string) {
	deleteRules, _ := c.ingress2Rule.Get(ingressName)

	// delete rules from ruleMap
	for _, rule := range deleteRules {
		rule := rule.(Rule)
		rules, ok := c.RuleMap[rule.GetHost()][rule.GetPath()]
		if !ok {
			continue
		}
		c.RuleMap[rule.GetHost()][rule.GetPath()] = delRule(rules, ingressName)
		if len(c.RuleMap[rule.GetHost()][rule.GetPath()]) == 0 {
			delete(c.RuleMap[rule.GetHost()], rule.GetPath())
		}
		if len(c.RuleMap[rule.GetHost()]) == 0 {
			delete(c.RuleMap, rule.GetHost())
		}
	}

	c.ingress2Rule.RemoveAll(ingressName)
}

func (c *httpBaseCache) put(rule Rule) error {
	if _, ok := c.RuleMap[rule.GetHost()]; !ok {
		c.RuleMap[rule.GetHost()] = make(map[string][]Rule)
	}

	for i, r := range c.RuleMap[rule.GetHost()][rule.GetPath()] {
		if annotations.Equal(rule.GetAnnotations(), r.GetAnnotations()) {
			// all conditions are same, oldest rule is valid
			if rule.GetCreateTime().Before(r.GetCreateTime()) {
				log.Log.V(0).Info("rule is overwritten by elder ingress", "ingress", r.GetIngress(), "host", r.GetHost(), "path", r.GetPath(), "old-ingress", rule.GetIngress())

				c.ingress2Rule.Remove(rule.GetIngress(), c.RuleMap[rule.GetHost()][rule.GetPath()][i])
				c.RuleMap[rule.GetHost()][rule.GetPath()][i] = rule
				c.ingress2Rule.Put(rule.GetIngress(), rule)
				return nil
			} else if rule.GetCreateTime().Equal(r.GetCreateTime()) {
				return nil
			} else {
				return fmt.Errorf("ingress [%s] conflict with existing %s, rule [host: %s, path: %s]", rule.GetIngress(), r.GetIngress(), rule.GetHost(), rule.GetPath())
			}
		}
	}
	c.ingress2Rule.Put(rule.GetIngress(), rule)
	c.RuleMap[rule.GetHost()][rule.GetPath()] = append(c.RuleMap[rule.GetHost()][rule.GetPath()], rule)

	return nil
}

func delRule(ruleList []Rule, ingress string) []Rule {
	var result []Rule
	for _, rule := range ruleList {
		if rule.GetIngress() != ingress {
			result = append(result, rule)
		}
	}
	return result
}

func checkHost(host string) error {
	// wildcard hostname: started with "*." is allowed
	if strings.Count(host, "*") > 1 || (strings.Count(host, "*") == 1 && !strings.HasPrefix(host, "*.")) {
		return fmt.Errorf("wildcard host[%s] is illegal, should start with *. ", host)
	}
	return nil
}

func checkPath(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("path is not set")
	}

	if strings.ContainsAny(path, "*") {
		return fmt.Errorf("path[%s] is illegal", path)
	}
	return nil
}
