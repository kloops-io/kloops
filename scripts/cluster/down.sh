#!/bin/bash

set -e

CLUSTER_NAME=kloops

function verify_supported() {
  if ! type "kind" > /dev/null 2>&1; then
    echo "kind is required"
    exit 1
  fi
}

function message() {
    echo "--------------------------------------------------------------------------------------------------------------"
    echo "$1"
    echo "--------------------------------------------------------------------------------------------------------------"
}

function delete_kind_cluster() {
    message "Deleting kind cluster $CLUSTER_NAME ..."
    kind delete cluster --name $CLUSTER_NAME
}

function cluster_down() {
    message "Cluster down !"
}

set -u

while [[ $# -gt 0 ]]; do
    case $1 in
        '--name')
            shift
            CLUSTER_NAME="${1}"
            ;;
        *)
            echo "ERROR: Unknown option $1"
            exit 1
            ;;
    esac
    shift
done

set +u

verify_supported
delete_kind_cluster
cluster_down