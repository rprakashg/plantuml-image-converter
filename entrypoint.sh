#!/bin/sh

CMD=$1

echo "Command :" $CMD

case $CMD in
    "start")
        echo "Starting Golang application"
        exec /app/plantuml-image-conversion
    ;;
esac
