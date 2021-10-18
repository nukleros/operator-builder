#!/bin/sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run ./cmd/operator-builder completion "$sh" >"completions/operator-builder.$sh"
done
