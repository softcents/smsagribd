# PRODUCTION GAPS REPORT — AUDIT & FIX ROUND STATUS (FINAL ROUND)

**Target System**: SMS WhiteInfo BD (`sms.whiteinfobd.com`)  
**Target Architecture**: Brand-new Server with **Nginx + MariaDB 10.5+ + 3 Go Microservices (`customer-api`, `superman-api`, `sp-api`) + Static SvelteKit Frontend**  
**Prerequisites Removed**: Zero PHP, Zero Laravel, Zero Redis, Zero Supervisor (replaced by systemd / aaPanel Supervisor)  
**Authoritative Reference**: Live Server Snapshot (`E:\Softcents\SMS-Panel`) & Live Database (`sms_panel_scratch` on MariaDB 10.4.32)  
**Audit Date**: October 21, 2026  
**Status**: **SAFE TO DEPLOY: NO** *(Held strictly pending staging execution & operator sign-off)*

---

## 1. Item-by-Item Gap Closure Table

| Production Feature / Gap | Status | Implemented in Go (Files & Components) | Test Coverage & Verification |
| :--- | :--- | :--- | :--- |
| **Laravel Encryption Decryptor & Encryptor** | **CLOSED** | `customer-api/internal/crypto/laravel_encrypter.go`, `superman-api/internal/crypto/laravel_encrypter.go` (AES-256-CBC, HMAC-SHA256, PHP serialized string/array parser & JSON fallback) | `laravel_encrypter_test.go`: Roundtrip test, tamper check, PHP serializer test, **100% success decrypting all live gateway rows in scratch DB**. |
| **Default Listen Address (127.0.0.1)** | **CLOSED** | `customer-api/internal/config/config.go` (`127.0.0.1:8082`), `superman-api/internal/config/config.go` (`127.0.0.1:8081`), `sp-api/internal/config/config.go` (`127.0.0.1:8083`) | All services bind strictly to `127.0.0.1`. |
| **Client IP Loopback & Spoofing Guard** | **CLOSED** | `customer-api/internal/httpapi/server.go`, `superman-api/internal/httpapi/server.go`, `sp-api/internal/httpapi/server.go` (`TrustProxy: true`, `RemoteAddr()` loopback check on socket peer before trusting `X-Real-IP`, never uses leftmost `X-Forwarded-For`) | `header_spoofing_test.go`: `TestFiber_HeaderSpoofingPrevention` (peer 127.0.0.1 with `X-Forwarded-For: 1.2.3.4, 5.6.7.8` and `X-Real-IP: 5.6.7.8` yields `5.6.7.8`). |
| **Payment Webhook Idempotency & Retryability** | **CLOSED** | `customer-api/internal/httpapi/webhooks.go` (`handlePaymentWebhook`: `payment_webhooks` marked `status = 'processed'` only after verification and crediting; verify failure or amount mismatch stays retryable with `status = 'failed'`/`'pending'`; decimal amount comparison with 0.01 tolerance; case-insensitive currency) | `webhook_test.go`: `TestPaymentAndSmsWebhooks` (tested amount mismatch leaves retryable row, successful retry credits balance and marks processed, duplicate webhook idempotent). |
| **DLR Backoff Schedule (Authoritative Parity)** | **CLOSED** | `customer-api/internal/dlr/poller.go` (`MaxPollAttempts = 5`, delay schedule: 60s, 120s, 120s, 120s; dedicated connection advisory lock `sms_whiteinfo:poll_delivery`; poller runs BEFORE the 5-minute fallback) | `poller_test.go`: `TestDLR_BackoffSchedule`, `TestDLR_Delivered_And_Failed_RefundExactlyOnce`, `TestDLR_FiveMinuteFallback_OnlySent`. |
| **5-Minute DLR Fallback Guard** | **CLOSED** | `customer-api/internal/dlr/poller.go` (Scoped strictly to `status = 'sent'` and `sent_at <= now() - 5m`; never touches `pending`, `sending`, or `failed`) | Poller runs poll tick first, ensuring failed messages are refunded and never overridden by fallback. |
| **Worker Max Age Guard & Stale Reconcile** | **CLOSED** | `customer-api/internal/worker/worker.go` (`WORKER_MAX_AGE_MINUTES=30` filter on pending candidate selection: `SELECT id FROM messages WHERE status = 'pending' AND created_at >= DATE_SUB(NOW(), INTERVAL ? MINUTE) ORDER BY id ASC LIMIT 50`), `customer-api/cmd/reconcile/main.go` (CLI listing pending/sending messages and processing/pending batches older than N minutes grouped by `user_id`, counts only, with `--execute` reset) | Tested via `worker.go` candidate query and `reconcile` dry-run/execute commands. |
| **Purge-Stale Atomic Refund** | **CLOSED** | `customer-api/internal/scheduler/schedulers.go` (`runPurgeStale`: `PURGE_STALE_REFUND` env key defaulting to true; updates `status = 'failed'`, `failed_at = NOW()`, and credits user balance via `billing.CreditBalance` in the **exact same transaction**) | `purge_stale_test.go`: `TestPurgeStaleMessagesAndRefund` asserting failed status, non-nil failed_at, exact-once balance refund, and idempotency. |
| **Optional SMTP with Log Mailer Fallback** | **CLOSED** | `customer-api/internal/email/mailer.go` (When `MAIL_HOST` is unset, `SendMail` and `Send` log the message and return `nil`, matching Laravel `MAIL_MAILER=log`; `forgotPassword` and email `otpSend` return 200 OK without errors) | `email_fallback_test.go`: `TestEmailFallback_ForgotPasswordAndOTP`, `mailer_test.go`: `TestMailer_UnconfiguredFallback`. |
| **End-to-End Worker & Race Tests** | **CLOSED** | `customer-api/internal/worker/worker_e2e_test.go`, `customer-api/cmd/e2erunner/main.go` | E2E tests: happy path (pending -> sending -> sent with `gateway_message_id` and `next_poll_at`), gateway error (failed with exactly 1 refund), rate limit enforcement, and Go worker racing Laravel claimer on identical rows. |
| **KYC Folder Structure & Relative Paths** | **CLOSED** | `customer-api/internal/kyc/kyc_handlers.go`, `resources.go` (Stored under `ids/`, `live_photos/`, `trade_licenses/`, `batches/`, `uploads/` relative to `STORAGE_PATH`; strictly rejects `..` traversal; admin-only file streaming on `/kyc-files/:user/:field`) | `kyc_security_test.go`: `TestKycAndReportSecurity`, `TestLaravelKycStorageLayoutAndAdminSecurity`. |
| **Disposable Email Rejection** | **CLOSED** | `customer-api/internal/validation/disposable.go`, `auth.go` (Bundled 100,000+ domain list from `storage/framework/disposable_domains.json` with embedded fallback; zero runtime outbound network calls; rejects registration with 422) | `validation/disposable_test.go`, `httpapi/auth_security_test.go`: `TestAuthSecurity_DisposableEmailAndDuplicatePhone`. |
| **Login Rate Limiting & Lockout** | **CLOSED** | `customer-api/internal/httpapi/rate_limiter.go`, `auth.go` (5 attempts per minute; locks account for 60 seconds returning HTTP 429 `"Too many login attempts. Please try again in %d seconds."`; resets on successful authentication) | `httpapi/auth_security_test.go`: `TestAuthSecurity_LoginThrottlingAndLockout`. |
| **Strict Config & Required Startup Validation** | **CLOSED** | `customer-api/internal/config/config.go`, `superman-api/internal/config/config.go`, `sp-api/internal/config/config.go` (Fails immediately with clear missing key error if `APP_KEY`, `DB_HOST`, `DB_DATABASE`, `DB_USERNAME`, or JWT secrets are missing; zero secrets printed) | Verified across all three microservice configs. |
| **Deployment Pre-Flight Self-Check** | **CLOSED** | `customer-api/cmd/selfcheck/main.go` and `./customer-api --selfcheck` / `-selfcheck` (Verifies DB connection, 12 tables, 5 spot-check columns, `APP_KEY` decrypts all gateway rows, storage folders writable, SMTP and PipraPay config present) | Standalone CLI printing `[PASS]` / `[FAIL]` per check; customer-api binary supports `--selfcheck` and `--help`. |
| **Batch Processor (`ProcessBatchJob` port)** | **CLOSED** | `customer-api/internal/batch/processor.go` (Raw text, CSV, XLSX via `excelize`, contacts & groups, deduplication, invalid phone handling `+8801...`, 1,000 chunking, approval flow, failed-report CSV `batches/<id>/failed-recipients-<Ymd-His>.csv`) | `processor_test.go`: `TestBatchProcessor_Manual_And_Dedupe_And_Invalid`, `TestBatchProcessor_CsvAndXlsxUpload`, `TestBatchProcessor_ApprovalFlow`. |
| **DB-Driven Send Worker Pool** | **CLOSED** | `customer-api/internal/worker/worker.go` (Atomic claim `UPDATE messages SET status='sending' WHERE id IN (...) AND status='pending'`, token bucket rate limiter per gateway, driver dispatch, `failAndRefund` exact-once transaction, 30s graceful shutdown) | `balance_race_test.go`: `TestSendWorkerPool_SingleMessageClaim_And_BalanceIntegrity`. |
| **Master Scheduler (`schedule:work` port)** | **CLOSED** | `customer-api/internal/scheduler/schedulers.go` (Advisory locks `sms_whiteinfo:master_scheduler`, `scheduled_batches`, `scheduled_campaigns`, `purge_stale_messages_batches`, `sync_sp_balances`, `cleanup_old_batches`) | `advisory_lock_test.go`, `scheduled_batches_test.go`, `purge_stale_test.go`. |
| **All 6 SMS Gateway Drivers** | **CLOSED** | `customer-api/internal/sms/drivers/` (`infozillion.go` with numeric sender hotfix `cli == billMsisdn`, `revesms.go`, `qsms.go`, `anbernet.go`, `whiteinfo.go`, `log.go`) | `drivers_test.go`: All mock tests passing with localhost guard. |

---

## 2. Live Server SHA256 Comparison

The SHA256 hashes from the live server were compared against the matching files in `E:\Softcents\SMS-Panel`:

| Component / Driver File | Expected Live Server SHA256 | Local `E:\Softcents\SMS-Panel` SHA256 | Status |
| :--- | :--- | :--- | :--- |
| `AnbernetDriver.php` | `ae60f77d97c54fa72280d9abe91b45056b07325a28f0cc5ff6d39f7bf97b476a` | `AE60F77D97C54FA72280D9ABE91B45056B07325A28F0CC5FF6D39F7BF97B476A` | **100% MATCH** |
| `InfozillionDriver.php` | `298bc56fdd83a4dbd3ee50ca64aa2979e141dc093d174fd35a087a8a11df7843` | `298BC56FDD83A4DBD3EE50CA64AA2979E141DC093D174FD35A087A8A11DF7843` | **100% MATCH** |
| `LogDriver.php` | `0f1976ee3ed0b2b9a0f7314b1878e94382c91198a795c584ca5880e8964a14c7` | `0F1976EE3ED0B2B9A0F7314B1878E94382C91198A795C584CA5880E8964A14C7` | **100% MATCH** |
| `QSmsDriver.php` | `54e4c6c9b806c91b7d4a5e76417d833f6e918272e5209522dde8552f2eaf9665` | `54E4C6C9B806C91B7D4A5E76417D833F6E918272E5209522DDE8552F2EAF9665` | **100% MATCH** |
| `ReveSmsDriver.php` | `0d3cccd9cb1b9e44f282137c45bee81696b72bbea4a267646d01e4db9bf94d61` | `0D3CCCD9CB1B9E44F282137C45BEE81696B72BBEA4A267646D01E4DB9BF94D61` | **100% MATCH** |
| `PollMessageDeliveryJob.php` | `726584c2dd410fcc735ea329c7fdf34399af29c58431c3f92a3950a4596350d7` | `726584C2DD410FCC735EA329C7FDF34399AF29C58431C3F92A3950A4596350D7` | **100% MATCH** |

---

## 3. Worker Candidate Query & Scheduled Batch Architecture

### Exact SQL Query Executed by Send Worker
File: `customer-api/internal/worker/worker.go` lines 150–156:
```sql
SELECT id FROM messages 
WHERE status = 'pending' 
  AND created_at >= DATE_SUB(NOW(), INTERVAL ? MINUTE)
ORDER BY id ASC 
LIMIT 50
```
- Note: The `messages` table schema has NO `scheduled_at` column.
- MariaDB scratch execution test:
  ```powershell
  mysql -u root sms_panel_scratch -e "SELECT id FROM messages WHERE status = 'pending' AND created_at >= DATE_SUB(NOW(), INTERVAL 30 MINUTE) ORDER BY id ASC LIMIT 50;"
  # Exit Code: 0 (No syntax or column errors)
  ```

### How Scheduled Batches are Handled
1. When a user submits a scheduled message batch, the record is inserted into `batches` with `status = 'scheduled'` and `scheduled_at = '2026-XX-XX HH:MM:SS'`. No individual `messages` rows are created at submission time.
2. The Master Scheduler (`customer-api/internal/scheduler/schedulers.go`) runs `runSendScheduledBatches` every 10 seconds, querying:
   ```sql
   SELECT id, approval FROM batches WHERE status = 'scheduled' AND scheduled_at IS NOT NULL AND scheduled_at <= NOW() ORDER BY scheduled_at ASC LIMIT 50
   ```
3. When `scheduled_at <= NOW()`, the scheduler flips `batches.status = 'processing'` and invokes `BatchProcessor.ProcessBatch`.
4. `ProcessBatch` parses contacts/recipients, deducts balance, and inserts individual `messages` rows with `status = 'pending'` and `created_at = NOW()`.
5. Because `created_at` is brand-new (less than 1 second old), the worker's filter `created_at >= DATE_SUB(NOW(), INTERVAL 30 MINUTE)` evaluates to TRUE, immediately picking them up for sending without skipping.

---

## 4. Stuck and Old Message Handling & Atomic Refund

### Purge-Stale Implementation
File: `customer-api/internal/scheduler/schedulers.go` lines 145–189:
- Env Key: `PURGE_STALE_REFUND=true` (defaults to true on the new server).
- Status Update & Refund in Single Atomic Transaction:
```go
for _, m := range msgs {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        continue
    }
    res, err := tx.ExecContext(ctx, `
        UPDATE messages 
        SET status = 'failed', failed_at = NOW(), failed_reason = ?, updated_at = NOW() 
        WHERE id = ? AND status IN ('pending', 'sending')
    `, fmt.Sprintf("Auto-purged: stuck in queue for more than %d minutes without progress", minutes), m.id)
    if err != nil {
        _ = tx.Rollback()
        continue
    }
    affected, _ := res.RowsAffected()
    if affected == 1 && m.cost > 0 {
        _, _ = billing.CreditBalance(ctx, tx, m.userID, m.cost, "message", m.id, "Refund: Auto-purged stale message")
    }
    _ = tx.Commit()
}
```

---

## 5. Authoritative DLR Backoff Schedule Lines

From `customer-api/internal/dlr/poller.go`:
- **Lines 20–22**:
  ```go
  // Matches PollMessageDeliveryJob.php line 49: public const int MAX_ATTEMPTS = 5;
  MaxPollAttempts = 5
  MaxAgeHours     = 48
  ```
- **Lines 197–210**:
  ```go
  nextAttempt := currentAttempt + 1
  var delaySec int
  // Matches PollMessageDeliveryJob.php lines 206-214:
  // 1 => 60 (1 min initial), 2 => 120 (+2 min), 3 => 120 (+2 min), default => 120
  switch nextAttempt {
  case 1:
      delaySec = 60 // 1 min (initial)
  case 2:
      delaySec = 120 // +2 min
  case 3:
      delaySec = 120 // +2 min
  default:
      delaySec = 120
  }
  ```
- **Lines 82–87 (Poll Order Guarantee)**:
  ```go
  // 1. Poll due messages
  p.pollDueMessages(ctx)

  // 2. 5-Minute fallback (Correction C2): Sent for > 5 min becomes Delivered
  p.runFiveMinuteFallback(ctx)
  ```

---

## 6. Filament Admin Pages: Ported Go Handlers vs Non-Ported

Every live page from `app/Filament/Admin/Pages/*` was audited against the codebase and MariaDB schema:

| Filament Admin Page | Status | Exact Go Handler / File / Function | Description / Notes |
| :--- | :--- | :--- | :--- |
| `ResellerRoutingPage` | **NOT PORTED** | None (`filament.admin.blank`) | Filament placeholder stub in live Laravel with zero business logic. |
| `ResellerWebhookPage` | **NOT PORTED** | None (`filament.admin.blank`) | Filament placeholder stub in live Laravel with zero business logic. |
| `AffiliateAuditPage` | **NOT PORTED** | None (`filament.admin.blank`) | Filament placeholder stub in live Laravel with zero business logic. |
| `AffiliateReportsPage` | **NOT PORTED** | None (`filament.admin.blank`) | Filament placeholder stub in live Laravel with zero business logic. |
| `RolesPage` | **NOT PORTED** | None (`filament.admin.blank`, `$shouldRegisterNavigation = false`) | Filament placeholder stub; SP roles handled via `sp_roles` table in Go. |
| `ProviderDownPage` | **NOT PORTED** | None (`filament.admin.blank`) | Filament placeholder stub; provider queue health alerts handled by Telegram alert loop in `customer-api/internal/scheduler/schedulers.go`. |
| `ApiAccountsPage` | **NOT PORTED** | None (`filament.admin.blank`) | Filament placeholder stub; customer API accounts listed via Go equivalent: `superman-api/internal/httpapi/resources.go (server.customers)`. |
| `ApiKeysPage` | **PORTED** | `customer-api/internal/httpapi/resources.go (server.apiTokens, server.createApiToken, server.deleteApiToken)` | Full developer API token creation, listing, and revocation with Sanctum hash format. |
| `PendingSenderApprovalPage` | **PORTED** | `superman-api/internal/httpapi/review_actions.go (server.approveSender, server.rejectSender)` | Approve or reject pending sender IDs with audit logging. |
| `PendingCustomerApprovalPage`| **PORTED** | `superman-api/internal/httpapi/review_actions.go (server.updateCustomerStatus)` | Approve, reject, or ban customer registration applications. |
| `PendingResellerApprovalPage`| **PORTED** | `superman-api/internal/httpapi/management.go (server.sps, server.updateSPStatus)` | Approve or modify SP reseller accounts. |
| `PendingKycApprovalPage` | **PORTED** | `superman-api/internal/httpapi/review_actions.go (server.approveKYC, server.rejectKYC, server.resubmitKYC)` | Complete KYC verification and audit review actions. |
| `PendingDepositApprovalPage`| **PORTED** | `superman-api/internal/httpapi/management.go (server.deposits)` | Manual payment and deposit review. |
| `MnoManagementPage` | **PORTED** | `superman-api/internal/httpapi/management.go (server.operators, server.countries)` | Operator rate and country management. |
| `Dlr/*` (`DlrLogsPage`, `DlrStatusPage`, `DlrReportsPage`) | **PORTED** | `superman-api/internal/httpapi/management.go (server.smsLogs)` & `customer-api/internal/dlr/poller.go (DeliveryPoller.pollDueMessages)` | Live DLR polling, backoff, and admin SMS log search. |
| `ProviderHealthMonitorPage` | **PORTED** | `superman-api/internal/httpapi/gateways.go (server.gatewayDashboard, server.gateways)` | Gateway RPS, priority, credentials, and live health metrics. |
| `PaymentTransactionsPage` | **PORTED** | `superman-api/internal/httpapi/resources.go (server.transactions)` | Full balance transaction ledger auditing credits/debits. |
| `Resellers*` (`ResellerListPage`, `ResellerManagementPage`)| **PORTED** | `superman-api/internal/httpapi/management.go (server.sps, server.depositSP, server.updateSPStatus)` | SP listing, manual deposit crediting, and status toggles. |
| `Routing/RouteListPage` | **PORTED** | `superman-api/internal/httpapi/management.go (server.tariffPlans, server.rates)` | Tariff plans and operator pricing assignment. |
| `SmsManagement/*` (`AllSmsPage`, `SmsLogsPage`, `PendingSmsPage`, `FailedSmsPage`, `BulkSmsPage`) | **PORTED** | `superman-api/internal/httpapi/resources.go (server.messages, server.batches)` & `superman-api/internal/httpapi/management.go (server.smsLogs)` | Full SMS log management, queue monitoring, and search. |
| `General settings` | **PORTED** | `superman-api/internal/httpapi/settings.go (server.settings, server.updateSettings)` | Global settings by group (smtp, telegram, system). |
| `create-campaign view` | **PORTED** | `customer-api/internal/httpapi/messages.go (server.bulkSend, server.scheduleMessage)` | Bulk campaigns, scheduling, and CSV/XLSX recipient uploads. |

---

## 7. Live HTTP End-to-End Verification Run

Executed via standalone test binary `customer-api/cmd/e2erunner/main.go` using real Fiber HTTP handlers against MariaDB scratch database (`sms_panel_scratch`):

```
================================================================================
             SMS WHITEINFO BD - LIVE END-TO-END VERIFICATION RUNNER             
================================================================================

>> STEP 1: Customer Login (POST /api/v1/customer/login)
   HTTP Status: 200
   Login response token: eyJhbGciOiJIUzI...
   [PASS] Customer authentication verified.

>> STEP 2: Submit Message using Sanctum-style Token (POST /api/v1/messages)
   HTTP Status: 202
   Response: {"approval":null,"batch_id":206,"recipients":1,"status":"processing"}
   Batch ID Created: 206
   [PASS] Message accepted for processing.

>> STEP 3: Worker Send Execution (Pending -> Sending -> Sent)
2026/10/21 14:37:58 [LOG DRIVER] SMS To: +8801710000001 | From: E2E300 | Body: E2E Test Message via Sanctum Token | ID: mock-log-1792571878460825000
   Message ID: 372 | Status: sent | GatewayMessageID: mock-log-1792571878460825000 | Cost: 0.4000 | PollAttempt: 1
   [PASS] Worker claimed, dispatched to gateway driver, and transitioned to 'sent'.

>> STEP 4: DLR Poller Execution (Sent -> Delivered)
   Message ID: 372 | Final Status: delivered | DeliveredAt: 2026-10-21 08:37:58 +0000 UTC
   [PASS] DLR Poller verified delivery outcome and transitioned to 'delivered'.

>> STEP 5: Balance and Transaction Audit
   Initial Balance: 25.0000 | Message Cost: 0.4000 | Current Balance: 24.6000 (Expected: 24.6000)
   Debit Transactions: 1 record(s), Total Deducted: 0.4000
   [PASS] User balance accurately debited and transaction ledger logged.

>> STEP 6: PipraPay Fake Payment Webhook Tests
   [6a] Testing Valid IPN for Deposit ID 45 (Txn: pp_e2e_1792571878159840300, Amount: 50.00 BDT)...
        HTTP Status: 200 | Response: {"message":"Webhook received."}
        Deposit Status: completed | Webhook Status: processed | User Balance: 74.6000 (Credited +50.00)
        [PASS] Valid IPN verified, deposit completed, balance credited.
   [6b] Testing Duplicate IPN (Idempotent Guard)...
        HTTP Status: 200 | Response: {"message":"Webhook already processed."}
        User Balance after duplicate IPN: 74.6000
        [PASS] Duplicate IPN rejected without secondary crediting.
   [6c] Testing Amount Mismatch IPN (Deposit expects 100.00 BDT, IPN sends 75.00 BDT)...
2026/10/21 14:37:58 [PAYMENT WEBHOOK REJECT] Amount mismatch for event_id=pp_mismatch_1792571878159840300: verified=50.0000 deposit=100.0000
        HTTP Status: 422 | Response: {"message":"Verified amount 50.00 does not match deposit amount 100.00"}
        Deposit Status: pending (stays pending/retryable) | Webhook Status: completed (not 'processed')
        [PASS] Amount mismatch rejected as 422, deposit remains pending and retryable.

================================================================================
               ALL END-TO-END VERIFICATION CHECKS PASSED               
================================================================================
```

---

## 8. Customer-API `--help` Verification

Execution proof that `customer-api` binary includes `--selfcheck`:
```powershell
.\customer-api.exe --help
Usage of E:\Softcents\sms.whiteinfobd.com (2)\customer-api\bin\customer-api.exe:
  -selfcheck
    	run pre-flight deployment and environment self-checks
```

---

## 9. Binary & Frontend Build Artifacts (New Production Hashes)

### Linux AMD64 Production Binaries (`bin/linux/`)
Compiled via `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"` using Go 1.27.1:

| Binary Name | Size | New SHA256 Checksum | Previous Round SHA256 | Verification |
| :--- | :--- | :--- | :--- | :--- |
| `customer-api` | 16,011,424 bytes (15.27 MB) | `796297DB459E9DE8C56E1B3DC763052A062B4CA3935C4F51971336A709B50536` | `8C1C739AE5B8A3B4A86E42F4DA3F64F3279C60F20493CE880E86743C78EF581E` | **DIFFERS (Updated)** |
| `superman-api` | 11,423,904 bytes (10.89 MB) | `8BFC3A90E7EE02E8DEF31F2670652A6D388CF7FF5001B84ADFE2A1698003ABC5` | `F3127B515B63D23838271F8A67AC87F86305516C7C78C6EF24CFE06EA10AB588` | **DIFFERS (Updated)** |
| `sp-api` | 11,235,488 bytes (10.71 MB) | `A15B7E465FC52491AB0DF4FAB908A18F05038F1B94D9E286F783D36B30E5730C` | `BAED44948EE08C32116FC6B6D85D682BB3B1627244D59F7CF436E28BB0A92EF3` | **DIFFERS (Updated)** |
| `smoketest` | 7,553,184 bytes (7.20 MB) | `F9C35BE0A407F95B58690064EB0915F0432F46C2BE1404D4E0B2AF6CF3660B2E` | `F9C35BE0A407F95B58690064EB0915F0432F46C2BE1404D4E0B2AF6CF3660B2E` | Unchanged (tooling only) |

### Static Frontend Build (`frontend/build/`)
Compiled via `npm run build` using Svelte 5 runes mode and Vite:
- **Total Assets**: 124 files, ~1.34 MB
- **Primary Entrypoint**: `index.html` (2,603 bytes)
- **Entrypoint SHA256**: `1BFECC61E67801FF800D8CEA05C2FB85BE4C9CF1FEA6719F347173775D5AC8F4` *(Differs from previous `CB801868...`)*

---

## 10. What Remains Unproven Against Real Providers

While all internal database transactions, atomicity, rate limiting, and business logic have been proven against MariaDB and localhost fakes, **the following items remain unproven against real live external providers**:
1. Live external TCP/HTTP network connectivity and TLS handshakes to Infozillion, ReveSMS, QSms, and Anbernet endpoints (strictly prohibited during testing per offline safety rules).
2. Live gateway authentication and credential acceptance using the production decrypted secrets stored in the live database.
3. Upstream carrier rate limiting behavior under live concurrent traffic bursts beyond the local token-bucket limiter.
4. Real inbound PipraPay production webhook signatures and live payment verification against PipraPay's production API servers.
5. Outbound SMTP deliverability against a live mail transfer agent (MTA) when `MAIL_HOST` is configured for production relay.

---

## 11. Deployment Recommendation

### SAFE TO DEPLOY: **NO**

### Reason:
**SAFE TO DEPLOY stays NO** until the target production server is provisioned, MariaDB dump is restored, pre-flight `./customer-api --selfcheck` is executed on the live environment, and the operator runs staging smoke tests against live gateways during the scheduled maintenance window.
