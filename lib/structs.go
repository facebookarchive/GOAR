// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package lib

import (
	"encoding/json"

	"github.com/facebookexperimental/GOAR/confighandler"
)

// Command abstracts command we want to execute
type Command struct {
	// Process name to be executied
	Cmd string
	// Arguments to be passed when executing process
	Args string
}

// Incident structure used for incidents recorded
type Incident struct {
	Rule         confighandler.Rule
	RawIncident  string
	Engine       string
	PreAudits    []*Command // All pre-audits needs to be successful for remediation to occur
	Remediations []*Command // Set of code that fix an issue or are part of a workflow (provision IP, discover neighbors, etc)
	PostAudits   []*Command // All post-audits needs to be succesul, should be code that makes sure everything is good after the execution
	Parameters   map[string]string
}

// IncidentToJSON converts the incident struct into a JSON string
func (inc Incident) IncidentToJSON() ([]byte, error) {
	return json.Marshal(inc)
}
