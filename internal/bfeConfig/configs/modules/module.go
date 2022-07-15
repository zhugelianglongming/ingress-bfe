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
	netv1 "k8s.io/api/networking/v1"
)

type BFEModuleConfig interface {
	UpdateIngress(ingress *netv1.Ingress) error
	DeleteIngress(ingressNamespace, ingressName string)
	Reload() error
	Name() string
}

type NewBFEModuleConfigFunc = func(version string) BFEModuleConfig

var initFuncList = []NewBFEModuleConfigFunc{NewRedirectConfig}

func InitBFEModules(version string) []BFEModuleConfig {
	var modules []BFEModuleConfig
	for _, initFunc := range initFuncList {
		modules = append(modules, initFunc(version))
	}
	return modules
}
