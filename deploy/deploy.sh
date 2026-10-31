#!/usr/bin/env bash
# ==============================================================================
# SMS WhiteInfo BD – Production Deployment & Rollback Script
# Target: Nginx + MariaDB + 3 Go Binaries + Static SvelteKit Frontend
# Usage:
#   ./deploy.sh install <tarball_or_dir>    Deploy new version with verification
#   ./deploy.sh rollback                   Revert to previous binary backup
#   ./deploy.sh verify                     Run selfcheck & health endpoints
# ==============================================================================

set -euo pipefail

BASE_DIR="/var/www/sms-panel"
BACKUP_DIR="${BASE_DIR}/backups/$(date +%Y%m%d_%H%M%S)"
SERVICES=("sms-customer-api" "sms-superman-api" "sms-sp-api")
SUPERVISOR_GROUP="sms-panel"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

err() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $*" >&2
}

# ------------------------------------------------------------------------------
# 1. Rollback Function
# ------------------------------------------------------------------------------
do_rollback() {
    log "Starting rollback to previous binaries..."

    for bin in "customer-api" "superman-api" "sp-api"; do
        BIN_PATH="${BASE_DIR}/${bin}/${bin}"
        BAK_PATH="${BASE_DIR}/${bin}/${bin}.bak"

        if [[ -f "${BAK_PATH}" ]]; then
            log "Restoring ${bin} from ${BAK_PATH}..."
            mv -f "${BAK_PATH}" "${BIN_PATH}"
            chmod 750 "${BIN_PATH}"
            chown www-data:www-data "${BIN_PATH}"
        else
            err "No backup found at ${BAK_PATH} for ${bin}!"
        fi
    done

    # Restart services
    log "Restarting services after rollback..."
    if command -v systemctl &>/dev/null && systemctl is-active --quiet sms-customer-api 2>/dev/null; then
        for svc in "${SERVICES[@]}"; do
            systemctl restart "${svc}"
        done
    elif command -v supervisorctl &>/dev/null; then
        supervisorctl restart "${SUPERVISOR_GROUP}:*"
    fi

    log "Rollback completed. Verifying health..."
    sleep 3
    if curl -s -f http://127.0.0.1:8082/health >/dev/null; then
        log "[SUCCESS] System healthy after rollback."
        exit 0
    else
        err "[CRITICAL] Health check failed even after rollback. Check logs immediately."
        exit 1
    fi
}

# ------------------------------------------------------------------------------
# 2. Verification Function
# ------------------------------------------------------------------------------
do_verify() {
    log "Running environment self-check..."
    if [[ -x "${BASE_DIR}/customer-api/selfcheck" ]]; then
        cd "${BASE_DIR}/customer-api"
        ./selfcheck || {
            err "selfcheck binary failed!"
            return 1
        }
    fi

    log "Checking health endpoints..."
    curl -s -f http://127.0.0.1:8082/health | grep -q '"database":"ok"' || {
        err "Customer API health check failed!"
        return 1
    }

    log "[PASS] All services verified healthy."
    return 0
}

# ------------------------------------------------------------------------------
# 3. Install / Deploy Function
# ------------------------------------------------------------------------------
do_install() {
    SOURCE_PATH="${1:-}"
    if [[ -z "${SOURCE_PATH}" ]]; then
        err "Usage: $0 install <release_directory_or_tarball>"
        exit 1
    fi

    log "Starting deployment from ${SOURCE_PATH}..."

    # Create backup directory
    mkdir -p "${BACKUP_DIR}"

    # Step 1: Verify SHA256 checksums if checksums.sha256 is present
    if [[ -f "${SOURCE_PATH}/checksums.sha256" ]]; then
        log "Verifying SHA256 checksums..."
        (cd "${SOURCE_PATH}" && sha256sum -c checksums.sha256) || {
            err "SHA256 checksum verification failed! Aborting deployment."
            exit 1
        }
        log "[PASS] Checksums verified successfully."
    fi

    # Step 2: Backup existing binaries
    log "Backing up current binaries..."
    for bin in "customer-api" "superman-api" "sp-api"; do
        BIN_PATH="${BASE_DIR}/${bin}/${bin}"
        if [[ -f "${BIN_PATH}" ]]; then
            cp -p "${BIN_PATH}" "${BASE_DIR}/${bin}/${bin}.bak"
            cp -p "${BIN_PATH}" "${BACKUP_DIR}/${bin}"
        fi
    done

    # Step 3: Install new binaries
    log "Installing new binaries..."
    for bin in "customer-api" "superman-api" "sp-api"; do
        SRC_BIN="${SOURCE_PATH}/${bin}"
        DEST_BIN="${BASE_DIR}/${bin}/${bin}"

        if [[ -f "${SRC_BIN}" ]]; then
            # Atomic install using temp file
            TMP_BIN="${DEST_BIN}.tmp"
            cp -f "${SRC_BIN}" "${TMP_BIN}"
            chmod 750 "${TMP_BIN}"
            chown www-data:www-data "${TMP_BIN}"
            mv -f "${TMP_BIN}" "${DEST_BIN}"
            log "Installed ${bin} -> ${DEST_BIN}"
        else
            err "Missing binary: ${SRC_BIN}"
            do_rollback
        fi
    done

    # Also install selfcheck and reconcile tools if present
    for tool in "selfcheck" "reconcile" "smoketest"; do
        if [[ -f "${SOURCE_PATH}/${tool}" ]]; then
            cp -f "${SOURCE_PATH}/${tool}" "${BASE_DIR}/customer-api/${tool}"
            chmod 750 "${BASE_DIR}/customer-api/${tool}"
            chown www-data:www-data "${BASE_DIR}/customer-api/${tool}"
        fi
    done

    # Step 4: Install frontend build if present
    if [[ -d "${SOURCE_PATH}/frontend_build" ]]; then
        log "Updating frontend static assets..."
        mkdir -p "${BASE_DIR}/frontend/build"
        rsync -a --delete "${SOURCE_PATH}/frontend_build/" "${BASE_DIR}/frontend/build/"
        chown -R www-data:www-data "${BASE_DIR}/frontend/build"
    fi

    # Step 5: Restart microservices
    log "Restarting services..."
    if command -v systemctl &>/dev/null && systemctl is-active --quiet sms-customer-api 2>/dev/null; then
        for svc in "${SERVICES[@]}"; do
            systemctl restart "${svc}"
        done
    elif command -v supervisorctl &>/dev/null; then
        supervisorctl restart "${SUPERVISOR_GROUP}:*"
    else
        log "Neither systemd nor supervisorctl detected. Please restart binaries manually."
    fi

    # Step 6: Post-deployment verification
    log "Verifying deployment health..."
    sleep 3
    if ! do_verify; then
        err "Deployment verification failed! Initiating automatic rollback..."
        do_rollback
    fi

    log "[SUCCESS] Deployment completed and verified healthy."
}

# ------------------------------------------------------------------------------
# Dispatcher
# ------------------------------------------------------------------------------
case "${1:-}" in
    install)
        shift
        do_install "$@"
        ;;
    rollback)
        do_rollback
        ;;
    verify)
        do_verify
        ;;
    *)
        echo "Usage: $0 {install <dir> | rollback | verify}"
        exit 1
        ;;
esac
