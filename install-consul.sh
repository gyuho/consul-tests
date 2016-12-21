#!/usr/bin/env bash
set -e

echo "Downloading Consul..."
rm -f /tmp/consul.zip
curl -sf -o /tmp/consul.zip https://releases.hashicorp.com/consul/0.7.2/consul_0.7.2_linux_amd64.zip

printf "\n"
echo "Extracting Consul..."
rm -f ${GOPATH}/bin/consul
unzip /tmp/consul.zip -d ${GOPATH}/bin
rm -f /tmp/consul.zip

printf "\n"
echo "Done!"
consul version

<<COMMENT
https://github.com/hashicorp/consul/releases
COMMENT
