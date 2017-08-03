#!/bin/bash

set -ve

if [ ! -z "$TRAVIS_TAG" ];then
	echo "the tag is $TRAVIS_TAG, will deploy...."
else
	echo "will not deploy..."
	exit 0
fi

make deploy

ghr -u ls0f -t $GITHUB_TOKEN -r cracker --replace  --debug $TRAVIS_TAG  dist/