#!/usr/bin/env bash

# Get the SHA of the working tree including all uncommitted changes.

# FOR DEBUG
# set -x

# Script should fail on errors, errors in pipes, or access to unset variables
set -o errexit -o pipefail -o nounset

# Create a temporary file path (just the path)
TF=$(mktemp)

# Remove the file (git requires an empty file)
rm -f $TF

# Copy the current git index to our temp location
cp .git/index $TF

# Stage all files using the temporary index (suppress warnings)
GIT_INDEX_FILE=$TF git add . &>/dev/null
# Generate the tree SHA from the temporary index
GIT_INDEX_FILE=$TF git write-tree
