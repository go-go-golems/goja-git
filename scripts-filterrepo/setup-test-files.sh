#!/bin/bash
# Setup script to create test files for filter-repo tests

set -e

echo "Setting up test files for filter-repo tests..."

# Test 1: Basic filter rename
mkdir -p /home/ubuntu/goja-git/test-filterrepo/src-basic/cmd/a
echo "file1 content" > /home/ubuntu/goja-git/test-filterrepo/src-basic/cmd/a/file1.txt
echo "file2 content" > /home/ubuntu/goja-git/test-filterrepo/src-basic/cmd/a/file2.txt
echo "file3 content" > /home/ubuntu/goja-git/test-filterrepo/src-basic/cmd/a/file3.txt

# Test 2: Extract to root
mkdir -p /home/ubuntu/goja-git/test-filterrepo/src-extract/cmd/a
echo "package main" > /home/ubuntu/goja-git/test-filterrepo/src-extract/cmd/a/main.go
echo "# Project A" > /home/ubuntu/goja-git/test-filterrepo/src-extract/cmd/a/README.md
echo "key: value" > /home/ubuntu/goja-git/test-filterrepo/src-extract/cmd/a/config.yaml

# Test 3: Prune empty
mkdir -p /home/ubuntu/goja-git/test-filterrepo/src-prune/other
mkdir -p /home/ubuntu/goja-git/test-filterrepo/src-prune/cmd/a
mkdir -p /home/ubuntu/goja-git/test-filterrepo/src-prune/docs
echo "other file 1" > /home/ubuntu/goja-git/test-filterrepo/src-prune/other/file1.txt
echo "other file 2" > /home/ubuntu/goja-git/test-filterrepo/src-prune/other/file2.txt
echo "package main" > /home/ubuntu/goja-git/test-filterrepo/src-prune/cmd/a/app.go
echo "package utils" > /home/ubuntu/goja-git/test-filterrepo/src-prune/cmd/a/utils.go
echo "# Documentation" > /home/ubuntu/goja-git/test-filterrepo/src-prune/docs/README.md

# Test 4: Deep paths
mkdir -p /home/ubuntu/goja-git/test-filterrepo/src-deep/pkg/api/v1/middleware
echo "package handler" > /home/ubuntu/goja-git/test-filterrepo/src-deep/pkg/api/v1/handler.go
echo "package types" > /home/ubuntu/goja-git/test-filterrepo/src-deep/pkg/api/v1/types.go
echo "package middleware" > /home/ubuntu/goja-git/test-filterrepo/src-deep/pkg/api/v1/middleware/auth.go
echo "package middleware" > /home/ubuntu/goja-git/test-filterrepo/src-deep/pkg/api/v1/middleware/logging.go

# Test 5: Monorepo
mkdir -p /home/ubuntu/goja-git/test-filterrepo/src-monorepo/projects/project-a
mkdir -p /home/ubuntu/goja-git/test-filterrepo/src-monorepo/projects/project-b
mkdir -p /home/ubuntu/goja-git/test-filterrepo/src-monorepo/docs
echo "# Monorepo" > /home/ubuntu/goja-git/test-filterrepo/src-monorepo/README.md
echo "package main" > /home/ubuntu/goja-git/test-filterrepo/src-monorepo/projects/project-a/main.go
echo "package utils" > /home/ubuntu/goja-git/test-filterrepo/src-monorepo/projects/project-a/utils.go
echo "key: value" > /home/ubuntu/goja-git/test-filterrepo/src-monorepo/projects/project-a/config.yaml
echo "package main" > /home/ubuntu/goja-git/test-filterrepo/src-monorepo/projects/project-b/main.go
echo "package server" > /home/ubuntu/goja-git/test-filterrepo/src-monorepo/projects/project-b/server.go
echo "# Architecture" > /home/ubuntu/goja-git/test-filterrepo/src-monorepo/docs/architecture.md

echo "✓ Test files created successfully"
