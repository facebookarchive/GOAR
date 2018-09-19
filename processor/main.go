// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"flag"

	"github.com/davecgh/go-spew/spew"

	"github.com/facebookexperimental/GOAR/confighandler"
	"github.com/golang/glog"
)

const (
	defaultQueueSize int  = 10
	blockingMode     bool = true
)

func main() {
	flag.Parse()

	glog.Infoln("[*] Starting syslog processor. To exit press CTRL+C")

	conf, err := confighandler.GetConfig("../config.yaml")
	if err != nil {
		glog.Exitf("Error opening config %s\n", err)
	}

	glog.Infoln("[*] Connection to queue server open")
	rules, err := confighandler.GetRules(conf.RulesFile)

	if err != nil {
		glog.Exitf("Error reading/parsing rules %s\n", err)
	}

	if glog.V(2) {
		glog.Infof("Read rules defined on %s: %v", conf.RulesFile, spew.Sdump(rules))
	}

	processor := NewProcessor()
	if err := processor.Connect(conf); err != nil {
		glog.Exitf("Error connecting, %v\n", err)
	}
	processor.Run(rules, blockingMode)
}
