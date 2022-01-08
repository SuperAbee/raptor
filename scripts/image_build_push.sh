#! /bin/bash

if [ -z $1 ]; then
    echo 'usage: ./image_build_push.sh $tag'
    echo 'example: ./image_build_push.sh latest'
    echo 'try again'
    exit
fi

cd ../

docker build -t superabee/raptor:$1 .

echo 'build ok'

docker push superabee/raptor:$1

echo 'push ok'