#!/bin/bash
# Modernized build script using Makefile

if [ "$1" = "public" ]; then
    echo "Building with local frontend compilation..."
    make build
else
    make build
fi