#!/usr/bin/env bash
set -e

PROJECT_DIR="/www/wwwroot/smsagri.softcents.com"
cd "$PROJECT_DIR"

export GIT_SSH_COMMAND="ssh -i /root/.ssh/id_ed25519 -o StrictHostKeyChecking=accept-new"

# Commit any uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Changes detected. Committing..."
    git add -A
    git commit -m "auto-sync: $(date '+%Y-%m-%d %H:%M:%S')"
fi

# Check if local is ahead of remote or if unpushed commits exist
git fetch origin main --quiet 2>/dev/null || true
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main 2>/dev/null || echo "")

if [ "$LOCAL" != "$REMOTE" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Pushing commits to GitHub..."
    git push origin main
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Pushed successfully to https://github.com/softcents/smsagribd"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Everything is in sync with GitHub."
fi
