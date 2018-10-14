#!/bin/sh

currdir=`dirname $0`
cd ${currdir}
echo "current dir=$currdir"

post=`utils/get-next-uwp.py 2>/dev/null`

echo "new post number=$post"
cd ../hugo

today=$(date +%Y-%m-%d)
hhmm=$(date +%H:%M)

outfile="./content/posts/podcast-$post.md"

echo '+++' > ${outfile}
echo 'layout: post' >> ${outfile}
echo "title = \"UWP - Выпуск $post\"" >> ${outfile}
echo "date: $today $hhmm" >> ${outfile}
echo 'comments: true' >> $outfile
echo 'categories: podcast' >> $outfile
echo "image = \"https://podcast.umputun.com/images/uwp/uwp$post.jpg\"" >> ${outfile}
echo "filename = \"ump_podcast${post}\"" >> ${outfile}
echo '+++' >> ${outfile}
echo ""  >> ${outfile}

echo "![](https://podcast.umputun.com/images/uwp/uwp${post}.jpg)\n" >> ${outfile}
echo - >>${outfile}
echo - >>${outfile}
echo - >>${outfile}
echo - >>${outfile}
echo - >>${outfile}
echo - >>${outfile}
echo - >>${outfile}
echo - >>${outfile}
echo "" >> ${outfile}
echo "[аудио](https://podcast.umputun.com/media/ump_podcast${post}.mp3)" >> ${outfile}
echo "<audio src=\"https://podcast.umputun.com/media/ump_podcast${post}.mp3\" preload=\"none\"></audio>" >> ${outfile}

st3 ${outfile} &
