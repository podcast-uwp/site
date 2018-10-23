#!/bin/sh
# this script runs from publisher directory

currdir=`dirname $0`
echo "current dir=$currdir"
cd ${currdir}

echo "generates site from ${currdir}"
num_before=`utils/get-next-uwp.py 2>/dev/null`

cd ..
git pull
git commit -am "auto episode after $num_before" && git push
#ssh podcast.umputun.com "cd /srv/podcast-uwp && git pull && docker-compose -f docker-compose-publisher.yml run --rm hugo"

num_after=`utils/get-next-uwp.py 2>/dev/null`
#if [[ $num_before != $num_after ]]
#then
#  link=`utils/get-last-uwp-link.py`
#  echo "will post new tweet for link $link"
  #./uwp.tweet "UWP $num_before $link #uwp"
#fi

echo "Done"