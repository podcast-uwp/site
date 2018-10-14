#!/bin/sh

wget https://podcast.umputun.com/categories/podcast/atom.xml -O /srv/octopress/public/atom-failback.xml
sed -i 's|https://podcast|http://podcast-failback|g' /srv/octopress/public/atom-failback.xml
