// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"flag"
	"fmt"

	"github.com/facebookexperimental/GOAR/confighandler"
	"github.com/golang/glog"
	syslog "gopkg.in/mcuadros/go-syslog.v2"

	"github.com/davecgh/go-spew/spew"
)

func formatSyslogListen(config *confighandler.Config) string {
	return fmt.Sprintf("%s:%s",
		config.SyslogIP,
		config.SyslogPort)
}

// Tail's by listening for syslog in the specificed IP/PORT, and sends
// the raw lines to rabbitmq on the configured queue (config.yaml - QUEUE_LOG)
func main() {

	flag.Parse()
	glog.Infoln("[*] Starting syslog tailer. To exit press CTRL+C")

	var err error

	configuration, err := confighandler.GetConfig("../config.yaml")
	fmt.Println(configuration)
	if err != nil {
		glog.Exitf("Error opening config: %s\n", err)
	}
	glog.Infof("Read of config complete configuring syslog server")
	logchannel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(logchannel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC5424)
	server.SetHandler(handler)
	listenStr := formatSyslogListen(&configuration)
	server.ListenUDP(listenStr)
	server.ListenTCP(listenStr)

	glog.Infof("Server configured. Attempting to start server")
	if glog.V(2) {
		glog.Infof("Syslog server parameters: %v", spew.Sdump(server))
	}

	server.Boot()

	tailer := &SyslogTailer{SyslogStream: logchannel}

	if err := tailer.PublishEndpoint.Connect(configuration); err != nil {
		glog.Exitf("Error connecting to remote queue: %s\n", err)
	}
	defer tailer.PublishEndpoint.Close()
	err = tailer.Tail()
	if err != nil {
		glog.Errorf("Error when tailing %s", err)
	}

}
