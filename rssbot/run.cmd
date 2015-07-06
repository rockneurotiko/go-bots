@echo off

SET BINARY=bin\rssbot_windows_amd64.exe

SET DBPATH=~\rss.db

SET ENVDIR=secrets.env

SET URL=""
@echo on

call %BINARY% --db=%DBPATH% --env=%ENVDIR% --deploy=%URL%
