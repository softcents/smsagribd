#!/usr/bin/env bash
set -e

PROJECT_DIR="/www/wwwroot/smsagri.softcents.com"
cd "$PROJECT_DIR"

export GIT_SSH_COMMAND="ssh -i /root/.ssh/id_ed25519 -o StrictHostKeyChecking=accept-new"

if [[ -n $(git status --porcelain) ]]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Changes detected. Syncing to GitHub..."
    git add -A
    git commit -m "auto-sync: $(date '+%Y-%m-%d %H:%M:%S')"
    git push origin main
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Pushed successfully to https://github.com/softcents/smsagribd"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Working tree clean. No changes to push."
fi
