// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package tailerhandler

import (
	"github.com/facebookexperimental/GOAR/confighandler"
	"github.com/facebookexperimental/GOAR/endpoints"
	"github.com/streadway/amqp"
)

// Event represents a type of each tailer event
type Event []byte

// Tailer defines an interface for a tailer.
type Tailer interface {
	Tail() error
	publish(event Event) error
}

// SyslogTailerQueue represents RabitMQ tailored for our tailer
type SyslogTailerQueue struct {
	OutQueue amqp.Queue
	endpoints.RabbitMQEndpoint
}

// Connect establishes and configures access to external queues. Returns error upon failure
func (endpoint *SyslogTailerQueue) Connect(config confighandler.Config) error {
	if err := endpoint.RabbitMQEndpoint.Connect(config); err != nil {
		return err
	}

	var err error
	endpoint.OutQueue, err = endpoint.Channel.QueueDeclare(
		config.QueueLog,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}
