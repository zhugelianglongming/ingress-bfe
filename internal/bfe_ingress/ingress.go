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

package bfe_ingress

import (
	"os"
	"syscall"
	"time"

	"github.com/baidu/go-lib/log"
	"github.com/bfenetworks/bfe/bfe_util/signal_table"
	networking "k8s.io/api/networking/v1beta1"

	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
)

type ingressList []*networking.Ingress

type BfeIngress struct {
	namespaces   []string
	labels       []string
	ingressClass string

	stopCh chan struct{}
}

func NewBfeIngress(namespaces, labels []string, ingressClass string) *BfeIngress {
	return &BfeIngress{
		namespaces:   namespaces,
		labels:       labels,
		ingressClass: ingressClass,
		stopCh:       make(chan struct{}),
	}
}

func (ing *BfeIngress) Start() {
	ingressesCh := make(chan ingressList, 1)
	client, _ := kubernetes_client.NewKubernetesClient()

	ing.initSignalTable()

	watcher := NewIngressWatcher(ing.namespaces, ing.labels, ing.ingressClass, client, ingressesCh, ing.stopCh)
	go watcher.Start()

	process := NewProcessor(client, ingressesCh, ing.stopCh)
	go process.Start()

	<-ing.stopCh
	log.Logger.Info("stop ingress")
}

func (ing *BfeIngress) Shutdown(sig os.Signal) {
	close(ing.stopCh)
	time.Sleep(time.Second)
}

func (ing *BfeIngress) initSignalTable() {
	/* create signal table */
	signalTable := signal_table.NewSignalTable()

	/* register signal handlers */
	signalTable.Register(syscall.SIGQUIT, ing.Shutdown)
	signalTable.Register(syscall.SIGTERM, signal_table.TermHandler)
	signalTable.Register(syscall.SIGHUP, signal_table.IgnoreHandler)
	signalTable.Register(syscall.SIGILL, signal_table.IgnoreHandler)
	signalTable.Register(syscall.SIGTRAP, signal_table.IgnoreHandler)
	signalTable.Register(syscall.SIGABRT, signal_table.IgnoreHandler)

	/* start signal handler routine */
	signalTable.StartSignalHandle()
}
