#!/usr/bin/env bash

for i in {1..10}; do
    echo "Starting messenger node instance $i"
    go run ./messenger &
done

wait
echo "All messenger nodes active."
