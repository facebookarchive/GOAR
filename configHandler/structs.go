// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package confighandler

// Rule represents a how a syslog matching rule should look like
type Rule struct {
	RuleName     string   `yaml:"RuleName"`
	AlertType    string   `yaml:"AlertType"`
	DeviceType   string   `yaml:"DeviceType"`
	Regex        string   `yaml:"Regex"`
	PreAudits    []string `yaml:"PreAudits"`
	Remediations []string `yaml:"Remediations"`
	PostAudits   []string `yaml:"PostAudits"`
}

// Config is a struc of configs :D
// Use for unmarshall the config file for basic parameters
// of the app
type Config struct {
	// Structure of the config
	QueueLog      string `yaml:"QUEUE_LOG"`
	QueueIncident string `yaml:"QUEUE_INCIDENT"`
	RabbitHost    string `yaml:"RABBITMQ_HOST"`
	RabbitPort    string `yaml:"RABBITMQ_PORT"`
	RabbitUser    string `yaml:"RABBITMQ_USER"`
	RabbitPass    string `yaml:"RABBITMQ_PASS"`
	SyslogFile    string `yaml:"LOGFILE"`
	RulesFile     string `yaml:"RULESFILE"`
	SyslogIP      string `yaml:"SYSLOG_LISTENIP"`
	SyslogPort    string `yaml:"SYSLOG_LISTENPORT"`
}
