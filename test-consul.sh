#!/usr/bin/env bash
set -e

echo "Installing consul-tests..."
go install -v

echo "Starting tests..."
nohup consul-tests -clients 1000 -endpoints localhost:8500 -key-size 8 -val-size 256 -writes 2000000 > ./consul-tests-result.txt 2>&1 &

