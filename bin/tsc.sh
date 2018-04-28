#!/bin/bash -eu

cd frontend/
node_modules/typescript/bin/tsc "$@"
