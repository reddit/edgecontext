#!/bin/sh

# This script can be used as a drop-in replacement of thrift compiler binary.
#
# You can override the version of thrift to use by using THRIFT_VERSION
# environment variable. Some of the valid versions are:
#
# - 0.13.0
# - 0.14.0
# - 0.14.1
#
# Check [1] for the full, up-to-date list.
#
# Note that due to docker's limitation on directory access, you cannot access
# (read or write) files outside of current directory ($PWD).
# For example, you can not do:
#
#     thrift-docker.sh ../path/to/file.thrift
#
# or
#
#     thrift-docker.sh -out ../output file.thrift
#
# [1]: https://github.com/orgs/reddit/packages/container/thrift-compiler/versions

DEFAULT_DOCKER_TAG="0.14.1"
DOCKER_TAG=${THRIFT_VERSION:-${DEFAULT_DOCKER_TAG}}
DOCKER_REPO=ghcr.io/reddit/thrift-compiler

docker run -v ${PWD}:/data/ --user "$(id -u):$(id -g)" ${DOCKER_REPO}:${DOCKER_TAG} "$@"
