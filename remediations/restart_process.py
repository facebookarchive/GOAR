#!/usr/bin/env python3

# Copyright (c) Facebook, Inc. and its affiliates.
# All rights reserved.

# This source code is licensed under the BSD-style license found in the
# LICENSE file in the root directory of this source tree.

import sys
import json
import logging


def restart_process():

    result = {
        "success": True,
        "passed": True,
        "result": "restart_process.py worked" 
    }
    print(json.dumps(result))
    logging.warning("Some stderr output")

if __name__ == "__main__":
    restart_process()