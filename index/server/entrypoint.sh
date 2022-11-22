#!/bin/bash
set -eux

# Check if devfile stacks and index.json exist
if [ ! -d "$DEVFILE_STACKS" ]; then
    echo "The container does not contain any devfile stacks in $DEVFILE_STACKS. Exiting..."
    exit 1
fi
if [ ! -e "$DEVFILE_INDEX" ]; then
    echo "The container does not contain an index.json at $DEVFILE_INDEX. Exiting..."
    exit 1
fi

# Start the index server
/registry/index-server
