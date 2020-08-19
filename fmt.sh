#!/bin/bash

go fmt *.go

files=$(find lib/ -type f -name "*.go")
for file in $files; do
    echo $file
    go fmt $file
done
