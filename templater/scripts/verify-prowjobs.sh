#!/usr/bin/env bash

set -e
set -u
set -o pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)

DIFF_LINE_COUNT=$(git diff --name-only $REPO_ROOT/jobs ':(exclude,top)*-1-23-*' | wc -l)
if [ $DIFF_LINE_COUNT -ne 0 ]; then
    CHANGED_FILES=$(git diff --name-only $REPO_ROOT/jobs ':(exclude,top)*-1-23-*')
    git diff $REPO_ROOT/jobs ':(exclude,top)*-1-23-*'
    echo "\n❌ Detected discrepancies between generated and expected Prowjobs!"
    echo "The following generated files need to be checked in:\n"
    echo "${CHANGED_FILES}\n" | tr ' ' '\n'
    exit 1
fi
