#!/usr/bin/python
# -*- coding: utf-8 -*-

import sys, os, string, time, smtplib, shutil, stat, urllib, glob

if __name__ == "__main__":
    line = os.popen("curl https://podcast.umputun.com/ | grep ump_podcast | head -n1").readline()
    link = line.split("\"")[1]
    print link
