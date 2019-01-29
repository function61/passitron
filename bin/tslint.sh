#!/bin/bash -eu

cd frontend/
tslint --project . "$@"
