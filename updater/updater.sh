#!/bin/sh
# this script runs outside of container, on host

cd /srv/podcast-uwp

git fetch;
LOCAL=$(git rev-parse HEAD);
REMOTE=$(git rev-parse @{u});

if [ $LOCAL != $REMOTE ]; then
    echo "$(date) git update detected"
    git pull origin master
    docker-compose -f docker-compose-publisher.yml run --rm hugo
    echo "$(date) update completed"
fi