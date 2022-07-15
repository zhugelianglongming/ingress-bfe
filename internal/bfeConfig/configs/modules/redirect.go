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

package modules

import (
	"fmt"

	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	"github.com/bfenetworks/bfe/bfe_modules/mod_redirect"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/condition"
	netv1 "k8s.io/api/networking/v1"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/cache"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
)

const (
	ConfigNameRedirect = "mod_redirect"
	RedirectRuleData   = "mod_redirect/redirect.data"
)

type ModRedirectConfig struct {
	version           string
	redirectRuleCache cache.Cache
	redirectConfFile  *mod_redirect.RedirectConfFile
}

func NewRedirectConfig(version string) BFEModuleConfig {
	return &ModRedirectConfig{
		version:           version,
		redirectRuleCache: cache.NewBaseCache(),
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
		r.redirectRuleCache.DeleteRulesByIngress(ingressName)
	}

	// check if the ingress has redirection annotations
	ok, err := annotations.HasRedirectAnnotations(ingress.Annotations)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// update cache
	if err := r.redirectRuleCache.UpdateByIngress(ingress); err != nil {
		return err
	}

	redirectConfFile, err := r.getRedirectTable()
	if err != nil {
		r.redirectRuleCache.DeleteRulesByIngress(ingressName)
		return err
	}
	r.redirectConfFile = redirectConfFile
	return nil
}

func (r ModRedirectConfig) getRedirectTable() (*mod_redirect.RedirectConfFile, error) {
	srcBasicRuleList, advancedRuleList := r.redirectRuleCache.GetRules()

	// remove the basicRule whose cluster is marked as AdvancedMode
	basicRuleList := make([]cache.Rule, 0)
	for _, rule := range srcBasicRuleList {
		if rule.GetCluster() == route_rule_conf.AdvancedMode {
			continue
		}
		basicRuleList = append(basicRuleList, rule)
	}
	ruleList := append(basicRuleList, advancedRuleList...)

	redirectRuleList := make(mod_redirect.RuleFileList, 0, len(ruleList))
	for _, rule := range ruleList {
		cmd, param, err := annotations.GetRedirectAction(rule.GetAnnotations())
		if err != nil {
			return nil, err
		}
		statusCode, err := annotations.GetRedirectStatusCode(rule.GetAnnotations())
		if err != nil {
			return nil, err
		}
		actions := &mod_redirect.ActionFileList{mod_redirect.ActionFile{
			Cmd:    &cmd,
			Params: []string{param},
		}}
		if err := mod_redirect.ActionFileListCheck(actions); err != nil {
			return nil, err
		}

		cond, err := condition.BuildCondition(rule.GetHost(), rule.GetPath(), rule.GetAnnotations())
		if err != nil {
			return nil, err
		}

		redirectRuleList = append(redirectRuleList, mod_redirect.RedirectRuleFile{
			Cond:    &cond,
			Actions: actions,
			Status:  &statusCode,
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

	r.redirectRuleCache.DeleteRulesByIngress(ingressName)
	r.getRedirectTable()
}

func (r *ModRedirectConfig) Reload() error {
	reload := false
	if *r.redirectConfFile.Version != r.version {
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
