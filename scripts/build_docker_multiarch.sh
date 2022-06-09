#!/usr/bin/env bash

# This script merges Coder Docker images of different architectures together
# into the archless image tag returned by ./image_tag.sh.
#
# Usage: ./build_docker_multiarch.sh [--version 1.2.3] [--push] image1:tag1 image2:tag2
#
# The supplied images must already be pushed to the registry or this will fail.
# Also, the source images cannot be in a different registry than the target
# image generated by ./image_tag.sh.
#
# If no version is specified, defaults to the version from ./version.sh.
#
# If the --push parameter is supplied, all supplied tags will be pushed.
#
# Returns the merged image tag.

set -euo pipefail
# shellcheck source=lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

version=""
push=0

args="$(getopt -o "" -l version:,push -- "$@")"
eval set -- "$args"
while true; do
    case "$1" in
    --version)
        version="$2"
        shift 2
        ;;
    --push)
        push=1
        shift
        ;;
    --)
        shift
        break
        ;;
    *)
        error "Unrecognized option: $1"
        ;;
    esac
done

# Remove the "v" prefix.
version="${version#v}"
if [[ "$version" == "" ]]; then
    version="$(execrelative ./version.sh)"
fi

if [[ "$#" == 0 ]]; then
    error "At least one argument must be provided to this script, $# were supplied"
fi

create_args=()
for image_tag in "$@"; do
    create_args+=(--amend "$image_tag")
done

output_tag="$(execrelative ./image_tag.sh --version "$version")"
log "--- Creating multi-arch Docker image ($output_tag)"
docker manifest create \
    "$output_tag" \
    "${create_args[@]}"

if [[ "$push" == 1 ]]; then
    log "--- Pushing multi-arch Docker image ($output_tag)"
    docker push "$output_tag"
fi

echo -n "$output_tag"
