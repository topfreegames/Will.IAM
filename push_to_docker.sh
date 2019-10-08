#!/bin/bash

set -euo pipefail

# The image build step occurs only when a build is triggered by a commit being merged into the "master" branch.
# Given that the "master" branch is protected, the only way to trigger a build from the "master" branch is when a PR
# is merged into it. This way, we avoid storing images from non-stable branches.
readonly DOCKER_HUB_REPO='tfgco/will-iam'

docker_tag_exists() {
    local repo="$1"
    local tag="$2"
    curl --silent -flSL "https://index.docker.io/v1/repositories/$repo/tags/$tag" > /dev/null
}

main() {
  local last_commit_sha
  last_commit_sha=$(git rev-parse --short HEAD)

  if docker_tag_exists "$DOCKER_HUB_REPO" "$TRAVIS_TAG"; then
      echo "An image with the version $TRAVIS_TAG already exists in Docker Hub. Please update your version.txt file and try again."
      exit 1
  fi

  docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
  docker build -t will-iam .
  docker tag will-iam "$DOCKER_HUB_REPO:$last_commit_sha"
  docker tag will-iam "$DOCKER_HUB_REPO:$TRAVIS_TAG"
  docker push "$DOCKER_HUB_REPO:$last_commit_sha"
  docker push "$DOCKER_HUB_REPO:$TRAVIS_TAG"
}

main "$@"
