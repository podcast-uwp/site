#!/bin/sh
# script runs inside hugo container
echo " === generate pages ==="
cd /srv/hugo
hugo
/srv/hugo/generate_rss.py
