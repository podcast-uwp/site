#!/usr/bin/python
# -*- coding: utf-8 -*-

import os

if __name__ == "__main__":
    line = os.popen("curl https://podcast.umputun.com/ | grep ump_podcast | head -n1").readline()
    num = int(line.split("ump_podcast")[1][:3])+1
    print num
