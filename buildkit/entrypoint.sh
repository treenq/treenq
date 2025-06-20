#!/bin/sh
set -e

# Start buildkitd with TCP address only
exec buildkitd --addr tcp://0.0.0.0:1234 --debug