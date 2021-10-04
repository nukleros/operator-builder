#!/usr/bin/env bash

COMMIT_MSG=`git log -n 1 --pretty=format:"%s"`

if [[ ! "$COMMIT_MSG" =~ ^((build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test|¯\\_\(ツ\)_\/¯)(\(\w+\))?(!)?(: (.*\s*)*))|(Merge (.*\s*)*)|(Initial commit$) ]]; then
    echo "commit message check failed:"
    echo
    echo "${COMMIT_MSG}"
    echo
    echo "message is not conventional commits format"
    echo "please see https://www.conventionalcommits.org/en/v1.0.0/#specification"

    exit 1
fi
