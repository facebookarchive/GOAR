// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"github.com/facebookexperimental/GOAR/tailerhandler"
	"github.com/golang/glog"
	"github.com/hpcloud/tail"
	"github.com/streadway/amqp"
)

// FileTailer represents a tailer that reads syslog events from a file and pushes them to a Rabbit queue.
type FileTailer struct {
	InputFileName   string
	PublishEndpoint tailerhandler.SyslogTailerQueue
}

// Tail reads data from given filename, line by line and pushes them out to external queue/component
func (tailer *FileTailer) Tail() error {
	tail, err := tail.TailFile(tailer.InputFileName, tail.Config{Follow: true})
	if err != nil {
		return err
	}

	for line := range tail.Lines {
		glog.V(2).Infof("Read line: %s ", line.Text)
		if err := tailer.publish(tailerhandler.Event(line.Text)); err != nil {
			glog.Errorf("Error publishing Event to an external queue, %s", err)
		}
	}
	return nil
}

// publish pushes event to a remote outgoing queue
func (tailer *FileTailer) publish(event tailerhandler.Event) error {

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
