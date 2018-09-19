// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/facebookexperimental/GOAR/tailerhandler"
	"github.com/golang/glog"
	"github.com/streadway/amqp"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

// SyslogTailer represents a tailer that reads syslog events from a syslog stream channel
type SyslogTailer struct {
	SyslogStream    syslog.LogPartsChannel
	PublishEndpoint tailerhandler.SyslogTailerQueue
}

// Tail reads data from a Syslog Channel line by line and pushes them out to external queue/component
func (tailer *SyslogTailer) Tail() error {
	glog.Infof("Syslog host tailer waiting for lines...")
	for logParts := range tailer.SyslogStream {
		if glog.V(2) {
			glog.Infof("Syslog received: %v", spew.Sdump(logParts))
		}
		if err := tailer.publish(tailerhandler.Event(fmt.Sprint(logParts))); err != nil {
			glog.Errorf("Error publishing Event to an external queue, %s", err)
		}
	}
	return nil
}

// publish pushes event to a remote outgoing queue
func (tailer *SyslogTailer) publish(event tailerhandler.Event) error {

	err := tailer.PublishEndpoint.Channel.Publish(
		"", // exchange
		tailer.PublishEndpoint.OutQueue.Name, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        event,
		})

	return err

}
