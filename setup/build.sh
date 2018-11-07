#!/bin/bash

set -e

KO_DOCKER_REPO=gcr.io/joe-does-knative ko apply -P -f service.yaml
