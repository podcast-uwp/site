#!/usr/bin/python3
# -*- coding: utf-8 -*-

import os

if __name__ == "__main__":
    line = os.popen(
        "curl -s https://podcast.umputun.com/ | grep ump_podcast | head -n1").readline()
    link = line.split("\"")[1]
    print(link)
