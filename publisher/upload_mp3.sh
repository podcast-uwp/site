#!/bin/bash
# this script runs from publisher directory

currdir=`dirname $0`
cd ${currdir}
echo "current dir=$currdir"

export LANG="en_US.UTF-8"
fname=`basename $1`
lftp="/usr/local/bin/lftp"
notif="/usr/local/bin/terminal-notifier"

episode=`echo $1 | sed -n 's/.*ump_podcast\(.*\)\.mp3/\1/p'`
image="umputun-art.png"

${notif} -title PodPrc -message "UWP detected #${episode}"
cd utils
./eyeD3.exec -v --remove-all --set-encoding=utf8 --album="Еженедельный подкаст от Umputun" --add-image=umputun-art.png:FRONT_COVER: --artist="Umputun" --track=${episode} --title="UWP Выпуск ${episode}" --year=$(date +%Y) --genre="Podcast" $1
${notif} -title PodPrc -message "UWP tagged"

cd ..

echo "upload to podcast.umputun.com"
${notif} -title PodPrc -message "upload started"
scp $1 podcast.umputun.com:/srv/podcast-uwp/var/media/${fname}
ssh podcast.umputun.com "chmod 644 /srv/podcast-uwp/var/media/${fname}"

echo "copy to hp-usrv archives"
${notif} -title PodPrc -message "copy to hp-usrv archives"
scp -P 2222 $1 umputun@192.168.1.24:/data/archive.rucast.net/uwp/media/

echo "upload to archive site"
scp $1 umputun@master.radio-t.com:/data/archive/uwp/media/${fname}

echo "remove old media files"
ssh podcast.umputun.com "/srv/podcast-uwp/publisher/cleanup.sh"

echo "all done for $fname"
${notif} -title PodPrc -message "all done for $fname"
