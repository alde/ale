#!/bin/bash

set -e

test -n "$DOCKER_USERNAME"
grep -E 'master|v[0-9.]+' > /dev/null <<< "$TRAVIS_BRANCH"

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
make docker
make publish-docker
