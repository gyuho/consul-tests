#!/usr/bin/env bash
set -e

for PORT in 8400 8500; do
    echo "SIGINT to" ${PORT}
    if [ -n "$(lsof -ti tcp:${PORT})" ]; then
        kill -2 $(lsof -ti tcp:${PORT})
        echo "Killed" ${PORT}
    else
        echo ${PORT} "has no processes"
    fi
done
