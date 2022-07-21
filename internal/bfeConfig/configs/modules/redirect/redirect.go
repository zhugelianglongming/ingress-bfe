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

package redirect

import (
	"fmt"

	"github.com/bfenetworks/bfe/bfe_modules/mod_redirect"
	netv1 "k8s.io/api/networking/v1"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
)

const (
	ConfigNameRedirect = "mod_redirect"
	RedirectRuleData   = "mod_redirect/redirect.data"
)

type ModRedirectConfig struct {
	version           string
	redirectRuleCache *redirectRuleCache
	redirectConfFile  *mod_redirect.RedirectConfFile
}

func NewRedirectConfig(version string) *ModRedirectConfig {
	return &ModRedirectConfig{
		version:           version,
		redirectRuleCache: newRedirectRuleCache(),
		redirectConfFile:  newRedirectConfFile(version),
	}
}

func newRedirectConfFile(version string) *mod_redirect.RedirectConfFile {
	ruleFileList := make([]mod_redirect.RedirectRuleFile, 0)
	productRulesFile := make(mod_redirect.ProductRulesFile)
	productRulesFile[configs.DefaultProduct] = (*mod_redirect.RuleFileList)(&ruleFileList)
	return &mod_redirect.RedirectConfFile{
		Version: &version,
		Config:  &productRulesFile,
	}
}

func (r *ModRedirectConfig) UpdateIngress(ingress *netv1.Ingress) error {
	if len(ingress.Spec.Rules) == 0 {
		return nil
	}

	ingressName := util.NamespacedName(ingress.Namespace, ingress.Name)
	if r.redirectRuleCache.ContainsIngress(ingressName) {
		r.redirectRuleCache.DeleteByIngress(ingressName)
	}

	// update cache
	if err := r.redirectRuleCache.UpdateByIngress(ingress); err != nil {
		return err
	}

	redirectConfFile, err := r.getRedirectTable()
	if err != nil {
		r.redirectRuleCache.DeleteByIngress(ingressName)
		return err
	}
	r.redirectConfFile = redirectConfFile
	return nil
}

func (r ModRedirectConfig) getRedirectTable() (*mod_redirect.RedirectConfFile, error) {
	ruleList := r.redirectRuleCache.GetRules()
	redirectRuleList := make(mod_redirect.RuleFileList, 0, len(ruleList))
	for _, rule := range ruleList {
		rule := rule.(*redirectRule)
		cond, err := rule.GetCond()
		if err != nil {
			return nil, err
		}
		redirectRuleList = append(redirectRuleList, mod_redirect.RedirectRuleFile{
			Cond:    &cond,
			Actions: rule.action,
			Status:  &(rule.statusCode),
		})
	}

	redirectConfFile := newRedirectConfFile(util.NewVersion())
	(*redirectConfFile.Config)[configs.DefaultProduct] = &redirectRuleList
	if err := mod_redirect.RedirectConfCheck(*redirectConfFile); err != nil {
		return nil, err
	}

	return redirectConfFile, nil
}

func (r *ModRedirectConfig) DeleteIngress(namespace, name string) {
	ingressName := util.NamespacedName(namespace, name)

	if !r.redirectRuleCache.ContainsIngress(ingressName) {
		return
	}

	r.redirectRuleCache.DeleteByIngress(ingressName)
	r.getRedirectTable()
}

func (r *ModRedirectConfig) Reload() error {
	reload := false
	if *r.redirectConfFile.Version != r.version {
		// TODO cache 更新到 redirectConfFile
		err := util.DumpBfeConf(RedirectRuleData, r.redirectConfFile)
		if err != nil {
			return fmt.Errorf("dump %s error: %v", RedirectRuleData, err)
		}
		reload = true
	}
	if reload {
		if err := util.ReloadBfe(ConfigNameRedirect); err != nil {
			return err
		}
		r.version = *r.redirectConfFile.Version
	}

	return nil
}

func (r ModRedirectConfig) Name() string {
	return "mod_redirect"
}
