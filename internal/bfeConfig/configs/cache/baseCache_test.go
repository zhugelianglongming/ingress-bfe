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

package cache

import (
	"testing"
	"time"
)

func Test_putBasic(t *testing.T) {

	cache := NewBaseCache()

	tests := []struct {
		name string
		args *BaseRule
		want bool
	}{
		{
			name: "rule1",
			args: NewBaseRule(
				"ingress1",
				"example.com",
				"/foo",
				nil,
				"svc1",
				time.Now(),
			),
			want: true,
		},
		{
			name: "rule2",
			args: NewBaseRule(
				"ingress2",
				"example.com",
				"/foo",
				nil,
				"svc1",
				time.Now(),
			),
			want: false,
		},
		{
			name: "rule2",
			args: NewBaseRule(
				"ingress3",
				"",
				"",
				nil,
				"svc1",
				time.Now(),
			),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.putRule(tt.args)
			if tt.want != (err == nil) {
				t.Errorf("routeRuleCache.Put() [%s] test fail", tt.name)
			}
		})
	}

}

func Test_putAdvanced(t *testing.T) {

	cache := NewBaseCache()

	tests := []struct {
		name string
		args *BaseRule
		want bool
	}{
		{
			name: "rule1",
			args: NewBaseRule(
				"ingress1",
				"example.com",
				"/foo",
				nil,
				"svc1",
				time.Now(),
			),
			want: true,
		},
		{
			name: "rule2",
			args: NewBaseRule(
				"ingress2",
				"example.com",
				"/foo",
				nil,
				"svc1",
				time.Now().Add(5*time.Second),
			),
			want: false,
		},
		{
			name: "rule3",
			args: NewBaseRule(
				"ingress3",
				"example.com",
				"/foo",
				map[string]string{"bfe.ingress.kubernetes.io/router.cookie": "aaa"},
				"svc2",
				time.Now(),
			),
			want: true,
		},
		{
			name: "rule3-1",
			args: NewBaseRule(
				"ingress3-1",
				"example.com",
				"/foo",
				map[string]string{"bfe.ingress.kubernetes.io/router.cookie": "aaa"},
				"svc2",
				time.Now().Add(5*time.Second),
			),
			want: false,
		},
		{
			name: "rule4",
			args: NewBaseRule(
				"ingress4",
				"example.com",
				"/foo",
				map[string]string{"bfe.ingress.kubernetes.io/router.header": "bbb"},
				"svc3",
				time.Now(),
			),
			want: true,
		},
		{
			name: "rule4-1",
			args: NewBaseRule(
				"ingress4-1",
				"example.com",
				"/foo",
				map[string]string{"bfe.ingress.kubernetes.io/router.header": "bbb"},
				"svc3",
				time.Now().Add(5*time.Second),
			),
			want: false,
		},
		{
			name: "rule5",
			args: NewBaseRule(
				"ingress5",
				"example.com",
				"/foo",
				map[string]string{"bfe.ingress.kubernetes.io/balance.weight": "ccc"},
				"svc2",
				time.Now().Add(5*time.Second),
			),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.putRule(tt.args)
			if tt.want != (err == nil) {
				t.Errorf("routeRuleCache.Put() test fail, name=%s", tt.name)
			}
		})
	}

	basicList, advancedList := cache.GetRules()

	if len(advancedList) != 3 || len(basicList) != 1 {
		t.Errorf("routeRuleCache.Put() test fail, size of rule should be 3, %d is returned", len(advancedList))
	}

}

func Test_sortRule(t *testing.T) {

	tests := []struct {
		name string
		args []*BaseRule
		want []string
	}{
		{
			name: "test1",
			args: []*BaseRule{
				NewBaseRule("ingress1", "example.com", "/foo", nil, "svc1", time.Now()),
				NewBaseRule("ingress2", "example.com", "/foo", map[string]string{"bfe.ingress.kubernetes.io/router.header": "aaa"}, "svc2", time.Now()),
				NewBaseRule("ingress3", "example.com", "/foo", map[string]string{"bfe.ingress.kubernetes.io/router.cookie": "bbb"}, "svc3", time.Now()),
				NewBaseRule("ingress4", "example.com", "/foo", map[string]string{"bfe.ingress.kubernetes.io/router.cookie": "ccc", "bfe.ingress.kubernetes.io/router.header": "ddd"}, "svc4", time.Now()),
				NewBaseRule("ingress5", "", "", nil, "svc5", time.Now()),
			},
			want: []string{"ingress4", "ingress3", "ingress2", "ingress1", "ingress5"},
		},
		{
			name: "test2",
			args: []*BaseRule{
				NewBaseRule("ingress1", "*.example.com", "/foo", nil, "svc1", time.Now()),
				NewBaseRule("ingress2", "aaa.example.com", "/foo", nil, "svc2", time.Now()),
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewBaseCache()

			for _, r := range tt.args {
				cache.putRule(r)
			}
			_, advancedList := cache.GetRules()

			for i, r := range advancedList {
				if r.GetIngress() != tt.want[i] {
					t.Errorf("sortAdvancedRule() test %s fail", tt.name)
				}
			}
		})
	}

}

func Test_deleteRule(t *testing.T) {

	tests := []struct {
		name string
		del  string
		args []*BaseRule
		want []string
	}{
		{
			name: "test1",
			del:  "ingress4",
			args: []*BaseRule{
				NewBaseRule("ingress1", "example.com", "/foo", nil, "svc1", time.Now()),
				NewBaseRule("ingress2", "example.com", "/foo", map[string]string{"bfe.ingress.kubernetes.io/router.header": "aaa"}, "svc2", time.Now()),
				NewBaseRule("ingress3", "example.com", "/foo", map[string]string{"bfe.ingress.kubernetes.io/router.cookie": "bbb"}, "svc3", time.Now()),
				NewBaseRule("ingress4", "example.com", "/foo", map[string]string{"bfe.ingress.kubernetes.io/router.cookie": "ccc", "bfe.ingress.kubernetes.io/router.header": "ddd"}, "svc4", time.Now()),
			},
			want: []string{"ingress3", "ingress2", "ingress1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewBaseCache()

			for _, r := range tt.args {
				cache.putRule(r)
			}
			cache.DeleteRulesByIngress(tt.del)

			_, advancedList := cache.GetRules()

			for i, r := range advancedList {
				if r.GetIngress() != tt.want[i] {
					t.Errorf("deleteRule() test %s fail", tt.name)
				}
			}
		})
	}

}
