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

# Start the registry viewer
npm start &

# Wait for server to start
until $(curl --output /dev/null --silent --head --fail http://localhost:3000/viewer); do
    printf '.'
    sleep 1
done

# Start the index server
/registry/index-server
