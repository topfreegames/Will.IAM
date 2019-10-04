#!/bin/bash

# The image build step occurs only when a build is triggered by a commit being merged into the "master" branch.
# Given that the "master" branch is protected, the only way to trigger a build from the "master" branch is when a PR
# is merged into it. This way, we avoid storing images from non-stable branches.
VERSION=$(cat ./version/version.go | grep "var VERSION" | awk ' { print $4 } ' | sed s/\"//g)

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker build -t will-iam .
docker tag will-iam:latest tfgco/will-iam:$VERSION.$TRAVIS_BUILD_NUMBER
docker tag will-iam:latest tfgco/will-iam:$VERSION
docker tag will-iam:latest tfgco/will-iam:latest
docker push tfgco/will-iam:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/will-iam:$VERSION
docker push tfgco/will-iam:latest

