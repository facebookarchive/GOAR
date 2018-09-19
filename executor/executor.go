// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"sync"
	"time"

	"github.com/golang/glog"

	"github.com/facebookexperimental/GOAR/confighandler"
	"github.com/facebookexperimental/GOAR/lib"
	"github.com/streadway/amqp"
)

// RemediationTimeout defines maximum number of seconds an audit or remediation can run
const RemediationTimeout = time.Second * 60

// Executor defines the Input endpoint and the location of the remediation / audits.
type Executor struct {
	InputEndpoint    InputEndpoint
	remediationsPath string
}

// NewExecutor is Executor's constructor function
func NewExecutor(remediationsPath string) *Executor {
	return &Executor{
		remediationsPath: remediationsPath,
	}
}

// Run starts listening on the input channel and processes incoming incidents, one by one.
// This is a blocking method.
func (executor *Executor) Run() {
	for job := range executor.InputEndpoint.DeliveryChannel {
		var incident lib.Incident
		if err := json.Unmarshal(job.Body, &incident); err != nil {
			glog.Errorf("Error decomposing job. discarding: %s\n", err)
			continue
		}
		executor.processIncident(&incident, &job)
	}
}

// Connect establishes the connection to remote queue endpoint and deals with all the possible
// failures during that phase
func (executor *Executor) Connect(conf confighandler.Config) error {
	return executor.InputEndpoint.Connect(conf)
}

// processIncident processes singular incident - it will concurently run all the pre checks,
// and if all of them success withing defined timeout
// the main execution will be called.
// After that, again, concurrently, post audits will be called.
func (executor *Executor) processIncident(incident *lib.Incident, job *amqp.Delivery) error {
	if executor.processPreAudits(incident) == false {
		glog.Warning("Preaudits failed, not continuing")
		job.Nack(false, true) // ack single job, but requeue
		return nil
	}

	if executor.processRemediation(incident) == false {
		glog.Warning("Remediation failed, not continuing")
		job.Nack(false, true) // ack single job, but requeue
		return nil
	}

	if executor.processPostAudits(incident) == false {
		glog.Warning("Postaudits failed, not continuing")
		job.Nack(false, true) // ack single job, but requeue
		return nil
	}

	if err := job.Ack(false); err != nil {
		glog.Warningf("Cannot ack job from the queue, %s\n", err)
		return err
	}
	return nil
}

// processPreAudits executes in parallel multiple 3-rd party scripts configured
// in the PreAudit argument of the processed Incident.
func (executor *Executor) processPreAudits(incident *lib.Incident) bool {
	return executor.execute(incident.PreAudits)
}

// processRemediation executes actual remediation.
func (executor *Executor) processRemediation(incident *lib.Incident) bool {
	return executor.execute(incident.Remediations)
}

// processPostAudits executes in parallel multiple 3-rd party scripts configured
// in the PostAudit argument of the processed Incident.
func (executor *Executor) processPostAudits(incident *lib.Incident) bool {
	return executor.execute(incident.PostAudits)
}

// execute takes care of the concurrent execution of underlying commands.
// For each command it will run a goroutine that spawns a separate process and runs binary
// (+args) from a lib.Command object.
// Upon the timout all processes will be killed.
// As a result a bool value is returned , true on success and false otherwise.
func (executor *Executor) execute(commands []*lib.Command) bool {

	results := make(chan *Result)
	var wg sync.WaitGroup
	wg.Add(len(commands))

	go func() {
		wg.Wait()
		close(results)
	}()

	ctx, ctxCancel := context.WithTimeout(context.Background(), RemediationTimeout)
	defer ctxCancel()

	for _, command := range commands {

		go func(ctx context.Context, command *lib.Command, results chan *Result) {
			defer wg.Done()

			result := &Result{
				ExitCode: tOK,
				Err:      nil,
			}

			cmd := exec.CommandContext(ctx, executor.remediationsPath+command.Cmd, command.Args)

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				result.ExitCode = tConfigErr
				result.Err = err
				results <- result
				return
			}

			stderr, err := cmd.StderrPipe()
			if err != nil {
				result.ExitCode = tConfigErr
				result.Err = err
				results <- result
				return
			}

			if err := cmd.Start(); err != nil {
				result.ExitCode = tConfigErr
				result.Err = err
				results <- result
				return
			}

			if err := json.NewDecoder(stdout).Decode(&result.ProcessOutput); err != nil {
				result.ExitCode = tOutputErr
				result.Err = err
				results <- result
				return
			}

			if buffer, err := ioutil.ReadAll(stderr); err != nil {
				result.ExitCode = tExecErr
				result.Err = err
				results <- result
			} else {
				result.ChildStdErr = buffer
			}

			if err := cmd.Wait(); err != nil {
				result.ExitCode = tOutputErr
				result.Err = err
				results <- result
				return
			}

			results <- result
		}(ctx, command, results)
	}

	// wait for the results

	for result := range results {
		if result.ExitCode != tOK {
			glog.Warningf("Error executing process: %s, exit code %v \n", result.Err, result.ExitCode)
			ctxCancel()
			return false
		}

		if result.ProcessOutput.Passed == false {
			glog.Warning("Process succeded but operation failed")
			ctxCancel()
			return false
		}

		glog.Infof("Processed incident: Exit code %d\nSuccess %v\nPassed %v\nResult %s\nStderr %v\n",
			result.ExitCode,
			result.ProcessOutput.Success,
			result.ProcessOutput.Passed,
			result.ProcessOutput.Result,
			string(result.ChildStdErr))

	}
	return true

}
