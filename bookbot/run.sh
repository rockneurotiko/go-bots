#!/usr/bin/env bash
BINARY=bin/bookbot_linux_amd64
BOOKSPATH=~/Libros
DBPATH=book.db
PWD=""

$BINARY --dir=$BOOKSPATH --db=$DBPATH --pwd=$PWD
