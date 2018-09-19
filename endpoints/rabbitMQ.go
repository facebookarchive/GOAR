// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package endpoints

import (
	"fmt"

	"github.com/streadway/amqp"

	"github.com/facebookexperimental/GOAR/confighandler"
)

// Endpoint defines the the behavior of an endpoint
type Endpoint interface {
	Connect() error
	Close() error
}

// RabbitMQEndpoint defines the struct used by a RabbitMQ endpoint
type RabbitMQEndpoint struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// Connect exeuctes a connectiong to a RabbitMQ instance with specified values
// in confighandler.Config Writes the connection and open channel to communicate with Rabbit
func (endpoint *RabbitMQEndpoint) Connect(config confighandler.Config) error {
	var err error
	connURL := formatConnectionURL(&config)
	endpoint.Connection, err = amqp.Dial(connURL)

	if err != nil {
		return err
	}

	endpoint.Channel, err = endpoint.Connection.Channel()

	return err
}

// Close takes case of closing connection to the RabbitMQ broker
func (endpoint *RabbitMQEndpoint) Close() error {
	if err := endpoint.Channel.Close(); err != nil {
		return err
	}
	return endpoint.Connection.Close()
}

// formatConnectionURL builds AMQP endpoint string used to connect to the queue
func formatConnectionURL(config *confighandler.Config) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		config.RabbitUser,
		config.RabbitPass,
		config.RabbitHost,
		config.RabbitPort)
}
