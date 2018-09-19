// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"github.com/facebookexperimental/GOAR/confighandler"
	"github.com/facebookexperimental/GOAR/endpoints"
	"github.com/streadway/amqp"
)

// InputEndpoint defines structure for establishing and maintaining
// connection with the input RabbitMQ
type InputEndpoint struct {
	DeliveryChannel <-chan amqp.Delivery
	AmqpQueue       amqp.Queue
	endpoints.RabbitMQEndpoint
}

// Connect establishes and sets RabbitMQ settings to tail from remote queue
func (endpoint *InputEndpoint) Connect(conf confighandler.Config) error {
	var err error

	if err = endpoint.RabbitMQEndpoint.Connect(conf); err != nil {
		return err
	}

	if err = endpoint.RabbitMQEndpoint.Channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		return err
	}

	endpoint.DeliveryChannel, err = endpoint.RabbitMQEndpoint.Channel.Consume(
		conf.QueueIncident, // queue
		"",                 // consumer
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)

	return err
}
