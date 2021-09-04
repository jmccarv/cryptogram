#!/bin/bash

while true; do
    inotifywait -e close_write *.go >/dev/null
    sleep .2
    echo "Rebuilding..."
    go build
    sleep 1
done
