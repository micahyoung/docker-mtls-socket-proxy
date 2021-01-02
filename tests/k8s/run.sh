#!/bin/bash
set -o errexit -o pipefail -o nounset
dir=$(realpath $(dirname $0))
kubectl apply -f ${dir}/deployment-linux.yml

# kubectl port-forward $(kubectl get pod --selector="app=test-app" --output jsonpath='{.items[0].metadata.name}') 2376:2376
