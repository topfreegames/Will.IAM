#!/bin/bash

# The image build step occurs only when a build is triggered by a commit being merged into the "master" branch.
# Given that the "master" branch is protected, the only way to trigger a build from the "master" branch is when a PR
# is merged into it. This way, we avoid storing images from non-stable branches.
VERSION=$(cat ./version/version.go | grep "var VERSION" | awk ' { print $4 } ' | sed s/\"//g)
VERSION_REGEX="^$VERSION$"
TAG=$(curl https://registry.hub.docker.com/v1/repositories/tfgco/khan/tags | \
    sed -e 's/[][]//g' -e 's/"//g' -e 's/ //g' | \
    tr '}' '\n'  | \
    awk -F: '{print $3}' | \
    grep "$VERSION_REGEX"
)

if [ "$TAG" = "$VERSION" ]; then
  echo "An image with this version already exists in Docker Hub. Please update your version.go file and try again."
fi

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker build -t will-iam .
docker tag will-iam:latest tfgco/will-iam:$VERSION
docker tag will-iam:latest tfgco/will-iam:latest
docker push tfgco/will-iam:$VERSION
docker push tfgco/will-iam:latest
