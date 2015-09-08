#!/usr/bin/env bash
BINARY="go run --race main.go"
CONFIG="./config.json"
$BINARY --config $CONFIG
