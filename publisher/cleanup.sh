#!/bin/sh

echo "cleanup old episodes"
keep=20

total=$(ls -1 /srv/podcast-uwp/var/media/ | wc -l)
echo "total episodes $total"

if [ "$total" -gt "$keep" ]
then
 to_remove=$((total-keep))
 echo "found old episodes. going to remove $to_remove"
 ls -1 /srv/podcast-uwp/var/media/ | head -n${to_remove} | xargs -I {} rm -f /srv/podcast-uwp/var/media/{}
else
 echo "nothing to delete"
fi
