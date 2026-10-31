#!/usr/bin/env bash
set -e

echo "Building Go Microservices..."
cd "$(dirname "$0")"

echo "[1/4] Building Master Orchestrator (smsagri)..."
go build -o smsagri main.go

echo "[2/4] Building customer-api..."
(cd customer-api && go build -o customer-api .)

echo "[3/4] Building superman-api..."
(cd superman-api && go build -o superman-api .)

echo "[4/4] Building sp-api..."
(cd sp-api && go build -o sp-api .)

echo "All binaries built successfully!"
