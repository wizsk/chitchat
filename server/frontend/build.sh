#!/bin/sh

while true;do
    pnpm build
    echo
    inotifywait -qe modify './src'
done
