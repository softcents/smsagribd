#!/usr/bin/env bash
set -e

PROJECT_DIR="/www/wwwroot/smsagri.softcents.com"
cd "$PROJECT_DIR"

export GIT_SSH_COMMAND="ssh -i /root/.ssh/id_ed25519 -o StrictHostKeyChecking=accept-new"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Pulling latest updates from GitHub..."
git pull origin main

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Rebuilding Go microservices..."
./build.sh

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Restarting orchestrator microservices..."
fuser -k 8081/tcp 2>/dev/null || true
fuser -k 8082/tcp 2>/dev/null || true
fuser -k 8083/tcp 2>/dev/null || true
fuser -k 8080/tcp 2>/dev/null || true
sleep 1
nohup /www/wwwroot/smsagri.softcents.com/smsagri > /www/wwwroot/smsagri.softcents.com/orchestrator.log 2>&1 &

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Update complete and services are live!"
