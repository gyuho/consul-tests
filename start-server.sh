#!/usr/bin/env bash
set -e

echo "Cleaning up Consul data directory..."
DATA_DIR=./data.consul
rm -rf ${DATA_DIR}

echo "Starting Consul..."
nohup consul agent -server -data-dir ${DATA_DIR} -http-port=8500 -advertise=127.0.0.1 -bind 0.0.0.0 -client 127.0.0.1 -bootstrap-expect 1 > ./consul-tests.log 2>&1 &

sleep 5s
echo "Consul members:"
consul members -rpc-addr=127.0.0.1:8400
curl localhost:8500/v1/catalog/nodes

sleep 3s
cat ./consul-tests.log

<<COMMENT
curl -X PUT -d 'bar' http://localhost:8500/v1/kv/foo
curl -v http://localhost:8500/v1/kv/foo

tail -f ./consul-tests.log

# For multi-node
consul agent -server -data-dir ${DATA_DIR} -bind LEADER_IP:8500 -client LEADER_IP:8500 -bootstrap-expect 3
consul agent -server -data-dir ${DATA_DIR} -bind IP_2:8500 -client IP_2:8500 -join LEADER_IP
consul agent -server -data-dir ${DATA_DIR} -bind IP_3:8500 -client IP_3:8500 -join LEADER_IP
COMMENT
