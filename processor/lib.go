// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/facebookexperimental/GOAR/confighandler"
	"github.com/facebookexperimental/GOAR/lib"
	"github.com/golang/glog"
)

// compileRegexRules compiles regular expressions based on configured rules.
func compileRegexRules(rules []confighandler.Rule) []*regexp.Regexp {
	regexList := make([]*regexp.Regexp, 0, len(rules))

	for _, rule := range rules {
		rgx, err := regexp.Compile(rule.Regex)
		if err != nil {
			glog.Errorf("Unable to compile %v regex %v\n", rule.Regex, err)
			continue
		}
		regexList = append(regexList, rgx)
	}
	return regexList
}

// FormatIncident returns an instance of Incident struct build upon rules, parameters, input message as well as engine
// With rule, parameters gathered, raw message and engine used to detect an 'incident'
// we create an Incident struct with all that information, and return it back to the caller.
func FormatIncident(rule confighandler.Rule, params map[string]string, msg string, engine string) lib.Incident {

	incident := lib.Incident{
		// Rule and RawIncident are mostly useful for troubleshooting
		// This will be used intensively in our future elastic logging
		Rule:        rule,   // Rule that triggered the event.
		RawIncident: msg,    // What triggered the event
		Parameters:  params, // Parameters that should be sent to a job
		Engine:      engine, // Engine that catched the event. Syslog, event, etc
	}

	parameters := formatCommandArguments(&params)

	incident.PreAudits = formatCommand(&rule.PreAudits, parameters)
	incident.Remediations = formatCommand(&rule.Remediations, parameters)
	incident.PostAudits = formatCommand(&rule.PostAudits, parameters)

	if glog.V(2) {
		glog.Infof("Formatted incident: %v", spew.Sdump(incident))
	}
	return incident
}

func formatCommandArguments(params *map[string]string) *string {
	arguments := make([]string, 0, len(*params))
	for param, value := range *params {
		arguments = append(arguments, fmt.Sprintf("--%s=%s", param, value))
	}
	result := strings.Join(arguments, " ")
	return &result
}

func formatCommand(commands *[]string, parameters *string) []*lib.Command {

	incidentCommands := make([]*lib.Command, 0, len(*commands))

	for _, command := range *commands {
		incidentCommands = append(incidentCommands, &lib.Command{Cmd: command, Args: *parameters})
	}

	return incidentCommands
}
