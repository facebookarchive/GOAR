// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"math"
	"sync"

	"github.com/facebookexperimental/GOAR/endpoints"

	"github.com/golang/glog"

	"github.com/streadway/amqp"

	"github.com/facebookexperimental/GOAR/confighandler"
	"github.com/facebookexperimental/GOAR/lib"
)

// Number of processors to run concurrently.
const defaultProcessorsNum = 10

// InputEndpoint represents all requires parts
// to maintain connection with remote queue (RabbitMQ in this case).
type InputEndpoint struct {
	DeliveryChannel <-chan amqp.Delivery
	AmqpQueue       amqp.Queue
	endpoints.RabbitMQEndpoint
}

// Connect establishes and maintains all the elements of the connection
// with the remote queue.
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
		conf.QueueLog, // queue
		"",            // consumer
		false,         // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)

	return err
}

// OutputEndpoint represents endpoint
// maintaining connection with the output external queue
type OutputEndpoint struct {
	AmqpQueue amqp.Queue
	endpoints.RabbitMQEndpoint
}

// Connect establishes and maintains all the elements of the connection
// with the remote queue.
func (endpoint *OutputEndpoint) Connect(conf confighandler.Config) error {

	var err error
	if err = endpoint.RabbitMQEndpoint.Connect(conf); err != nil {
		return err
	}

	endpoint.AmqpQueue, err = endpoint.Channel.QueueDeclare(
		conf.QueueIncident, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	return err

}

// Processor listens to incoming events coming from the
// input queue (typically Rabbit), matches execution with
// precomiles regex and pushes Incident objects to executors.
type Processor struct {
	InputEndpoint  InputEndpoint
	OutputEndpoint OutputEndpoint

	RawLogChannel   chan []byte
	IncidentChannel chan lib.Incident

	eventProcessors int
}

// Connect maintains connections for both input and output of the Processor.
func (processor *Processor) Connect(conf confighandler.Config) error {

	if err := processor.InputEndpoint.Connect(conf); err != nil {
		return err
	}

	// Output Endpoint
	return processor.OutputEndpoint.Connect(conf)
}

// NewProcessor configures and sets Processor object.
func NewProcessor() *Processor {
	return &Processor{
		eventProcessors: defaultProcessorsNum,
		RawLogChannel:   make(chan []byte),
		IncidentChannel: make(chan lib.Incident),
	}
}

// Run runs all the pieces, listening for logs in the input queue, applying
// rules passed as an argument and publishes incidents to output external queue
func (processor *Processor) Run(rules []confighandler.Rule, blocking bool) {

	processor.tailInput()
	processor.publishIncidents()

	if blocking {
		processor.processEvents(rules)
	} else {
		go processor.processEvents(rules)
	}

}

func (processor *Processor) tailInput() {
	glog.Info("Entering tailInput() goroutine")
	go func(in <-chan amqp.Delivery, out chan []byte) {
		for message := range in {
			glog.V(2).Infof("Received a message: %s", message.Body)
			out <- message.Body
			message.Ack(false)
		}
		glog.Info("Exiting tailInput() goroutine")
	}(processor.InputEndpoint.DeliveryChannel, processor.RawLogChannel)

}

func (processor *Processor) processEvents(rules []confighandler.Rule) {

	regexList := compileRegexRules(rules)

	var wg sync.WaitGroup
	wg.Add(processor.eventProcessors)

	for i := 0; i < processor.eventProcessors; i++ {
		go func(processor *Processor) {
			defer wg.Done()
			for msg := range processor.RawLogChannel {
				msgStr := string(msg)
				for id, reg := range regexList {
					if reg.MatchString(msgStr) {

						eventParameters := make(map[string]string)

						values := reg.FindStringSubmatch(msgStr)
						regexNames := reg.SubexpNames()
						length := int(math.Min(float64(len(values)), float64(len(regexNames))))
						// skip first element in both arrays as these are full lines, not split arguments/parameters
						for i := 1; i < length; i++ {
							eventParameters[regexNames[i]] = values[i]
						}

						processor.IncidentChannel <- FormatIncident(rules[id], eventParameters, msgStr, "SYSLOGPROC")
						// As soon as we manage to create an incident from matching
						// any of the rules we skip to the next possible regex.
						// that way we avoid too much processing and also creating multiple
						// incidents from a single syslog line.
						break
					}
				}
			}
		}(processor)
	}
	glog.Info("Processing events...")
	wg.Wait()
}

func (processor *Processor) publishIncidents() {

	go func(incidentChannel chan lib.Incident, queueName string) {
		var body []byte
		var err error
		for {
			msg := <-incidentChannel

			if body, err = msg.IncidentToJSON(); err != nil {
				glog.Errorf("Error marshaling incident to JSON: %s", err)
				continue
			}

			if err = processor.OutputEndpoint.Channel.Publish(
				"",        // exchange
				queueName, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        body,
				}); err != nil {
				glog.Errorf("Error publishing incident to the Executor's queue, %s", err)
			}
		}
	}(processor.IncidentChannel, processor.OutputEndpoint.AmqpQueue.Name)

}
