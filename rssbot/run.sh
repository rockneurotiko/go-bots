#!/usr/bin/env bash
BINARY=bin/rssbot_linux_amd64
DBPATH=rss.db
ENVDIR=secrets.env
URL=""

$BINARY --db $DBPATH --env $ENVDIR --deploy=$URL
