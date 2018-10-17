#!/usr/bin/env sh

echo "activate stats updater"
cp -fv /index.html /stats/index.html
/usr/sbin/vnstatd -n &

while :
do
    echo "$(date) update stats"
    /usr/bin/vnstati  -s -i eth0 -o /stats/vnstat.png
    /usr/bin/vnstati -h -c 15 -i eth0  -o /stats/vnstat_h.png
    /usr/bin/vnstati -m -i eth0 -o /stats/vnstat_m.png
    /usr/bin/vnstati -d -i eth0 -o /stats/vnstat_d.png
    /usr/bin/vnstati -t -i eth0 -o /stats/vnstat_t.png
    sleep 5m
done
