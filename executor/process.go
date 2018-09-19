// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

type exitCode int

const (
	tOK exitCode = iota
	tConfigErr
	tOutputErr
	tExecErr
)

// ProcessOutput structures
// output from underlying audits/remediations
type ProcessOutput struct {
	// Boolean value depicting if the Audit/Remediation worked.
	Success bool `json:"success"`
	// Passed describes in a true/false manner if the audit passed
	Passed bool `json:"passed"`
	// Result represents arbitrary result passed from
	// underlying audit process.
	Result string `json:"result"`
}

// Result represents
// result of executed command
type Result struct {
	// ExitCode has non zero values with different values dependin on when the execution failed.
	// Please examine exicCode type and constant values it provides.
	ExitCode exitCode
	// Copied stderr. This is where process should send logs/info to be human readable.
	ChildStdErr []byte
	// in case ExitCode is not equal to tOK a corresponding error should be set here.
	Err error
	// Instance of well defined output - an effect of working audit or remediation.
	ProcessOutput ProcessOutput
}
