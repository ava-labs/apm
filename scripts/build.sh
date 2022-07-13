#!/usr/bin/env bash
# Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
# See the file LICENSE for licensing terms.


set -o errexit
set -o nounset
set -o pipefail

if ! [[ "$0" =~ scripts/build.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

# Set default binary directory location
name="apm"

# Build the apm
mkdir -p ./build

echo "Building apm in ./build/$name"
go build -o ./build/$name ./main
