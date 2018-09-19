// Copyright (c) Facebook, Inc. and its affiliates.
// All rights reserved.

// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package confighandler

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// GetConfig readds locally defined yaml config file
// Unmarshall the file and return a Config struct to the caller
func GetConfig(path string) (Config, error) {
	var conf Config
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return conf, err
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

// GetRules reads the yaml file containing the rules and unmarshal the
// info into a Rule
func GetRules(path string) ([]Rule, error) {
	var rules []Rule
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return rules, err
	}
	err = yaml.Unmarshal(yamlFile, &rules)
	if err != nil {
		return rules, err
	}
	return rules, nil
}
