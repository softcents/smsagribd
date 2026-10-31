# SMS WhiteInfo BD – Production Migration & Cutover Runbook

**Target Architecture**: Single or Multi-Node Linux Server running **Nginx + MariaDB 10.5+ + 3 Go Microservices (`customer-api`, `superman-api`, `sp-api`) + Static SvelteKit Frontend**.  
**Prerequisites Removed**: Zero PHP, Zero Laravel, Zero Redis, Zero Supervisor (replaced by systemd units or aaPanel Supervisor entries).  
**Domain**: `sms.whiteinfobd.com`  
**Status**: **SAFE TO DEPLOY: NO** *(Held pending staging execution & operator sign-off)*

---

## 1. MariaDB Setup & Dump Restoration

### A. Create MariaDB Database & User
```bash
mariadb -u root -p << 'EOF'
CREATE DATABASE IF NOT EXISTS sms_panel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'sms_panel'@'127.0.0.1' IDENTIFIED BY 'STRONG_SECRET_PASSWORD';
GRANT ALL PRIVILEGES ON sms_panel.* TO 'sms_panel'@'127.0.0.1';
FLUSH PRIVILEGES;
EOF
```

### B. Restore Database Dump
```bash
# Decompress and import dump
zcat sms_whiteinfo_YYYY-MM-DD.sql.gz | mariadb -u sms_panel -p'STRONG_SECRET_PASSWORD' sms_panel
```

---

## 2. Storage Directory Restoration

The Go microservices expect unencrypted, plain image and CSV files at `STORAGE_PATH` (`/var/www/sms-panel/storage/app/private`).

```bash
# 1. Create target storage layout
mkdir -p /var/www/sms-panel/storage/app/private/{ids,live_photos,trade_licenses,batches,uploads}

# 2. Extract private archive from source server
tar -xzvf sms-live-private.tar.gz -C /var/www/sms-panel/storage/app/private/

# 3. Set strict file permissions (readable & writable by www-data)
chown -R www-data:www-data /var/www/sms-panel/storage
find /var/www/sms-panel/storage/app/private -type d -exec chmod 750 {} +
find /var/www/sms-panel/storage/app/private -type f -exec chmod 640 {} +
```

---

## 3. Row-Count & Data Parity Comparison

Execute the following sanity query on both Old (Source) and New (MariaDB) databases to confirm parity:

```sql
SELECT 'users' AS tbl, COUNT(*) AS rows_count FROM users
UNION ALL SELECT 'messages', COUNT(*) FROM messages
UNION ALL SELECT 'batches', COUNT(*) FROM batches
UNION ALL SELECT 'senders', COUNT(*) FROM senders
UNION ALL SELECT 'gateways', COUNT(*) FROM gateways
UNION ALL SELECT 'deposits', COUNT(*) FROM deposits
UNION ALL SELECT 'transactions', COUNT(*) FROM transactions
UNION ALL SELECT 'payment_webhooks', COUNT(*) FROM payment_webhooks
UNION ALL SELECT 'access_tokens', COUNT(*) FROM access_tokens
UNION ALL SELECT 'tariff_plans', COUNT(*) FROM tariff_plans
UNION ALL SELECT 'tariff_rates', COUNT(*) FROM tariff_rates
UNION ALL SELECT 'ip_whitelist_requests', COUNT(*) FROM ip_whitelist_requests;
```

---

## 4. Environment & Binary Configuration

### A. Place Environment Files
Copy templates from `deploy/*.env.template` to each respective service directory:
- `/var/www/sms-panel/customer-api/.env`
- `/var/www/sms-panel/superman-api/.env`
- `/var/www/sms-panel/sp-api/.env`

> [!IMPORTANT]
> The `APP_KEY` variable must be identical across all three services and match the legacy Laravel `APP_KEY` (`base64:...`).

### B. Execute Automated Pre-Flight Self-Check
Before starting any service, run `selfcheck`:

```bash
cd /var/www/sms-panel/customer-api
./selfcheck
```

Expected output:
```text
  [PASS] DB_CONNECTION: Connected to sms_panel@127.0.0.1:3306
  [PASS] DB_TABLES: All 12 required tables present
  [PASS] DB_COLUMNS: All 5 spot-check columns present
  [PASS] APP_KEY_DECRYPT: gateway 1 (infozillion) OK
  [PASS] APP_KEY_DECRYPT: gateway 2 (revesms) OK
  [PASS] APP_KEY_DECRYPT: gateway 3 (qsms) OK
  [PASS] APP_KEY_DECRYPT: gateway 4 (anbernet) OK
  [PASS] APP_KEY_DECRYPT: gateway 5 (whiteinfo) OK
  [PASS] APP_KEY_DECRYPT: gateway 6 (log) OK
  [PASS] STORAGE_FOLDERS: /var/www/sms-panel/storage/app/private exists, ids/ live_photos/ trade_licenses/ batches/ OK
  [PASS] SMTP_CONFIG: MAIL_HOST=... MAIL_PORT=...
  [PASS] PIPRAPAY_CONFIG: Keys present

  RESULT: All 13 checks PASSED
```

---

## 5. Staging Smoke Tests

Keep `WORKER_ENABLED=false` and `SCHEDULER_ENABLED=false` initially. Start all three Go binaries and test:

### A. Three-Portal Authentication
1. **Customer Portal**: Sign in via `POST /api/v1/customer/login` with customer credentials. Confirm JWT token and cookie issued.
2. **Superman Portal**: Sign in via `POST /api/v1/superman/login` with admin credentials. Confirm dashboard metrics return.
3. **SP Reseller Portal**: Sign in via `POST /api/v1/sp/login` with reseller credentials. Confirm sub-customer list returns.

### B. Single Test SMS via Smoketest Utility
```bash
cd /var/www/sms-panel/customer-api
./smoketest --to=8801700000000 --body="Staging test message"
```
Verify message row inserted with `status = 'sent'`, valid `gateway_message_id`, and `next_poll_at` set.

### C. DLR Polling Verification
Run the poller tick to verify gateway status fetch:
```bash
# Verify log output records status resolution to 'delivered'
```

### D. Sandbox Payment Deposit Webhook
Simulate a sandbox payment callback:
```bash
curl -X POST http://127.0.0.1:8082/webhooks/payment/piprapay \
  -H "Content-Type: application/json" \
  -d '{"status":"completed","pp_id":"test_deposit_001","amount":100.00,"currency":"BDT"}'
```
Verify:
1. HTTP 200 response.
2. `deposits.status` updated to `completed`.
3. `users.balance` credited by exactly 100.00.
4. `payment_webhooks.processed_at` populated.
5. Re-sending identical request returns `"Webhook already processed."` without double-crediting.

### E. KYC File Streaming Check
Request a KYC document as admin:
```bash
curl -i -H "Authorization: Bearer <ADMIN_JWT>" \
  http://127.0.0.1:8082/kyc-files/1/front_nid
```
Confirm HTTP 200 and image binary stream. Confirm unauthorized requests receive HTTP 401 or 403.

---

## 6. Enabling Background Workers & Schedulers

Once smoke tests pass:
1. In `/var/www/sms-panel/customer-api/.env`, update:
   ```ini
   WORKER_ENABLED=true
   SCHEDULER_ENABLED=true
   ```
2. Restart `sms-customer-api`:
   ```bash
   systemctl restart sms-customer-api
   ```
3. Check health and worker status:
   ```bash
   curl http://127.0.0.1:8082/health
   # Returns: {"database":"ok","in_flight":0,"last_loop":"2026-10-21T...","status":"ok"}
   ```

---

## 7. Production Cutover & Rollback Runbook

### Pre-Cutover (T-24 Hours)
1. Lower DNS TTL for `sms.whiteinfobd.com` to 300 seconds (5 minutes).
2. Deploy code, binaries, and static assets to the new server.
3. Validate staging self-check (`./selfcheck`) passes 100%.

### Cutover Window (Maintenance Window)
1. **Quiesce Old Server** (T-0):
   ```bash
   # On old server: stop web traffic and horizon queues
   systemctl stop nginx
   php artisan horizon:terminate || true
   supervisorctl stop all
   ```
2. **Delta Database Dump & Sync**:
   ```bash
   # On old server: dump delta records
   mysqldump --single-transaction --quick sms_panel | gzip > delta_cutover.sql.gz
   scp delta_cutover.sql.gz user@new-server:/tmp/
   
   # Sync any last-minute uploaded files
   rsync -avz /www/wwwroot/sms.whiteinfobd.com/storage/app/private/ \
     user@new-server:/var/www/sms-panel/storage/app/private/
   ```
3. **Restore Delta on New Server**:
   ```bash
   zcat /tmp/delta_cutover.sql.gz | mariadb -u sms_panel -p'STRONG_SECRET_PASSWORD' sms_panel
   ```
4. **Run Reconcile Probe**:
   ```bash
   /var/www/sms-panel/customer-api/reconcile --minutes=30 --dry-run=true
   ```
5. **Start Go Microservices & Nginx**:
   ```bash
   systemctl start sms-customer-api sms-superman-api sms-sp-api
   systemctl restart nginx
   ```
6. **Switch DNS Record**:
   Update `sms.whiteinfobd.com` A record to point to the New Server IP.
7. **Monitor In-Flight Traffic**:
   Watch `journalctl -u sms-customer-api -f` and verify incoming webhooks and message dispatch.

### Rollback Procedure (If Aborting Cutover)
1. **Revert DNS Record**: Re-point DNS A record to Old Server IP.
2. **Stop New Server**:
   ```bash
   systemctl stop nginx
   systemctl stop sms-customer-api sms-superman-api sms-sp-api
   ```
3. **Dump Incremental Rows from New Server** (if any traffic hit):
   Export any `messages`, `deposits`, or `transactions` created on the new server during the window and replay onto old DB.
4. **Re-Enable Old Server**:
   ```bash
   systemctl start nginx
   php artisan horizon
   supervisorctl start all
   ```
5. Restore DNS TTL to 86400 (24h).
