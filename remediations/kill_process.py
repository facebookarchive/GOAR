#!/usr/bin/env python3

# Copyright (c) Facebook, Inc. and its affiliates.
# All rights reserved.

# This source code is licensed under the BSD-style license found in the
# LICENSE file in the root directory of this source tree.

import sys
import json
import logging


def kill_process():

    result = {
        "success": True,
        "passed": True,
        "result": "Process is killed" 
    }
    print(json.dumps(result))
    logging.warning("kill_process.py: Some stderr output")
    logging.warning("I am sooo unsure")

if __name__ == "__main__":
    kill_process()