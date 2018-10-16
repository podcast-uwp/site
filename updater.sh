#!/bin/sh
echo "$(date) start updater"
cd /srv/podcast-uwp

while true
do
    git fetch;
    LOCAL=$(git rev-parse HEAD);
    REMOTE=$(git rev-parse @{u});

    if [ $LOCAL != $REMOTE ]; then
        echo "$(date) git update detected"
        git pull origin master
        docker-compose -f docker-compose-publisher.yml run --rm hugo
        echo "$(date) update completed"
    fi
    sleep 10
done
