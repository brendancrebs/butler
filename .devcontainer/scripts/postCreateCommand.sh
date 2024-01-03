#!/usr/bin/env bash
#Copyright (c) 2023 Schweitzer Engineering Laboratories, Inc.
#SEL Confidential

cat <<EOF >>~/.zshrc
# butler is an alias for running butler with go run
function butler {
  go run /workspaces/butler/code/main.go "\$@"
}
EOF