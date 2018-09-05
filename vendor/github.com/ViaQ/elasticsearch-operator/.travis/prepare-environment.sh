#!/bin/sh

set -x
set -o errexit
set -o nounset

sudo sysctl -w vm.max_map_count=262144

go get -u github.com/golang/dep/cmd/dep
dep ensure
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.9.0/bin/linux/amd64/kubectl
chmod +x kubectl && sudo mv kubectl /usr/local/bin/
curl -Lo minikube https://storage.googleapis.com/minikube/releases/v0.25.2/minikube-linux-amd64
chmod +x minikube && sudo mv minikube /usr/local/bin/

sudo minikube start --vm-driver=none --kubernetes-version=v1.9.0
go get -u github.com/operator-framework/operator-sdk/commands/operator-sdk
minikube update-context
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do
  sleep 1; done
