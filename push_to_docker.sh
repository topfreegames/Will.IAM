#!/bin/bash

set -euo pipefail

# The image build step occurs only when a build is triggered by a commit being merged into the "master" branch.
# Given that the "master" branch is protected, the only way to trigger a build from the "master" branch is when a PR
# is merged into it. This way, we avoid storing images from non-stable branches.
DOCKER_HUB_REPO="tfgco/will-iam"
LAST_COMMIT_SHA=$(git rev-parse --short HEAD)
VERSION=$(cat version.txt)

function docker_tag_exists() {
    curl --silent -f -lSL "https://index.docker.io/v1/repositories/$1/tags/$2" > /dev/null
}

if docker_tag_exists "$DOCKER_HUB_REPO" "$VERSION"; then
    echo "An image with the version $VERSION already exists in Docker Hub. Please update your version.txt file and try again."
    exit 1
fi

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker build -t will-iam .
docker tag will-iam "$DOCKER_HUB_REPO":"$LAST_COMMIT_SHA"
docker tag will-iam "$DOCKER_HUB_REPO":"$VERSION"
docker tag will-iam "$DOCKER_HUB_REPO":latest
docker push "$DOCKER_HUB_REPO":"$LAST_COMMIT_SHA"
docker push "$DOCKER_HUB_REPO":"$VERSION"
docker push "$DOCKER_HUB_REPO":latest
