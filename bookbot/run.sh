#!/usr/bin/env bash
BINARY=bin/bookbot_linux_amd64
BOOKSPATH=~/Libros
DBPATH=book.db
PWD=""
ENVDIR=secrets.env

$BINARY --dir=$BOOKSPATH --db=$DBPATH --pwd=$PWD --env=$ENVDIR
