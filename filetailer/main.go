// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"flag"

	"github.com/golang/glog"

	"github.com/facebookexperimental/GOAR/confighandler"
)

// Tail's from the configured file (config.yaml - LOGFILE) and send the new
// raw syslog-like to rabbitmq on the configured queue (config.yaml - QUEUE_LOG)
// This works mostly like `tailf -f` on a Linux, the difference is that lines
// are not actually printed, but send to a rabbitmq server for processing by
// other consumers
func main() {

	flag.Parse()
	glog.Infoln("[*] Starting log tailer. To exit press CTRL+C")

	var err error

	configuration, err := confighandler.GetConfig("../config.yaml")
	if err != nil {
		glog.Exitf("Error opening config: %s\n", err)
	}

	tailer := &FileTailer{InputFileName: configuration.SyslogFile}

	if err := tailer.PublishEndpoint.Connect(configuration); err != nil {
		glog.Exitf("Error connecting to remote queue: %s\n", err)
	}
	defer tailer.PublishEndpoint.Close()
	err = tailer.Tail()
	if err != nil {
		glog.Errorf("Error when tailing %s", err)
	}
}
