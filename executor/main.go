// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"flag"

	"github.com/facebookexperimental/GOAR/confighandler"
	"github.com/golang/glog"
)

var remediationsPath = flag.String("remediations",
	"../remediations/",
	"Path to the directory with your remediations scripts")
var configPath = flag.String("config", "../config.yaml", "Path to the configuration file")

func main() {
	flag.Parse()

	glog.Infof("Getting a config %s\n", *configPath)
	conf, err := confighandler.GetConfig(*configPath)
	if err != nil {
		glog.Exitf("Error while reading config file: %s\n", err)
	}
	engine := NewExecutor(*remediationsPath)

	if err := engine.Connect(conf); err != nil {
		glog.Exitf("Error while establishing connection: %v\n", err)
	}

	glog.Info("Running executor")
	engine.Run()
}
