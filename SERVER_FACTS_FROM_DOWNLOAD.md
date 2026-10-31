# Server Facts & Download Audit Report

**Date of Audit:** October 20, 2026  
**Folder Inspected:** `C:\Users\asus\Downloads\sms.whiteinfobd.com (1)`  
**Compared Against:** `E:\Softcents\SMS-Panel` (October 16, 2026 snapshot)  
**Archive Reference:** `C:\Users\asus\Downloads\sms.whiteinfobd.com (1).tar.gz` (563 MB)  
**Scratch Database:** `sms_panel_scratch` (restored from `sms_whiteinfo_2026-10-16.sql`)

---

## 1. Age Check & File Comparison

### Summary Verdict
> [!IMPORTANT]
> **`C:\Users\asus\Downloads\sms.whiteinfobd.com (1)` is OLDER than `E:\Softcents\SMS-Panel`.**
> - The codebase inside `C:\Users\asus\Downloads\sms.whiteinfobd.com (1)` (and its source archive `sms.whiteinfobd.com (1).tar.gz`) dates from **September 22, 2026 to October 09, 2026**.
> - The live snapshot `E:\Softcents\SMS-Panel` dates from **October 16, 2026 at ~03:36 AM** and contains subsequent production changes (e.g., the WhiteInfo gateway driver, campaigns supervisor, and production optimizations).

### Filesystem & Extraction State
- **Drive C Disk Space Status:** During inspection, Drive `C:\` was discovered to have **0 bytes free** (`Free = 0`).
- Because of this disk full condition, extraction of `sms.whiteinfobd.com (1).tar.gz` into `C:\Users\asus\Downloads\sms.whiteinfobd.com (1)` was abruptly interrupted midway at `node_modules`.
- Consequently, in the physical folder `C:\Users\asus\Downloads\sms.whiteinfobd.com (1)`, files following `node_modules` alphabetically (such as `routes/`, `storage/`, `tests/`, `public/`, `resources/`, and `.env`) are not present on disk.
- However, full analysis was performed across both the extracted files on disk and the complete archive `sms.whiteinfobd.com (1).tar.gz`.

### Files Differing Between the Two Folders
Between common files present in both folders (excluding `vendor/`, `node_modules/`, and `.git/`), exactly **54 files** differ in content and timestamps:

| File Path | `sms.whiteinfobd.com (1)` (Downloads) Modified Date | `SMS-Panel` (Oct 16 Snapshot) Modified Date | Functional Difference |
|---|---|---|---|
| `app/Enums/GatewayDriver.php` | 2026-09-22 01:34:51 | 2026-10-16 03:15:13 | `SMS-Panel` adds `case WhiteInfo = 'whiteinfo'` and enum match arm. Missing in Downloads. |
| `app/Enums/IpWhitelistStatus.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:27 | Code formatting / whitespace normalization. |
| `app/Enums/KycPermission.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:27 | Array key formatting / alignment. |
| `app/Filament/Admin/Pages/AuditLogs/IpWhitelistAuditPage.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:28 | Import cleanup (`KeyValueEntry`, `TextEntry`) and badge color map alignment. |
| `app/Filament/Admin/Pages/Dashboard.php` | 2026-09-25 01:39:53 | 2026-10-16 02:23:29 | Whitespace cleanup. |
| `app/Filament/Admin/Pages/Kyc/KycAuditPage.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:29 | Import cleanup and array alignment. |
| `app/Filament/Admin/Pages/Kyc/KycDashboardPage.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:30 | Removed unused `FormTextInput` import; query clone optimization. |
| `app/Filament/Admin/Pages/Kyc/KycPermissionsPage.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:30 | Typed schema parameter `Schema $schema` and actor ID assignment fix. |
| `app/Filament/Admin/Pages/Kyc/VerificationRequestsPage.php` | 2026-10-06 18:16:26 | 2026-10-16 02:23:30 | Removed unused `UserStatus` import. |
| `app/Filament/Admin/Pages/PaymentGateway/GlobalPaymentGatewaysPage.php` | 2026-10-06 18:48:11 | 2026-10-16 02:23:30 | String concatenation formatting. |
| `app/Filament/Admin/Pages/PaymentGateway/PaymentReportsPage.php` | 2026-10-06 19:18:16 | 2026-10-16 02:23:30 | Removed unused `SupermanNavigationGroup` and `UnitEnum` imports. |
| `app/Filament/Admin/Pages/PaymentGateway/PaymentTransactionsPage.php` | 2026-10-06 19:18:16 | 2026-10-16 02:23:30 | Removed unused `SupermanNavigationGroup` and `UnitEnum` imports. |
| `app/Filament/Admin/Pages/PaymentGateway/ReconciliationPage.php` | 2026-10-06 19:18:16 | 2026-10-16 02:23:30 | Removed unused `SupermanNavigationGroup` and `UnitEnum` imports. |
| `app/Filament/Admin/Pages/Providers/ProviderHealthPage.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:30 | Removed unused navigation group import. |
| `app/Filament/Admin/Pages/Reports/CustomerReportsPage.php` | 2026-09-25 02:37:45 | 2026-10-16 02:23:31 | Line ending / newline normalization. |
| `app/Filament/Admin/Pages/Reports/ProviderReportsPage.php` | 2026-09-25 02:40:42 | 2026-10-16 02:23:32 | Filament v5 icon and schema imports update. |
| `app/Filament/Admin/Pages/Routing/FailoverPage.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:33 | Removed unused navigation group import. |
| `app/Filament/Admin/Pages/Routing/RoutingRulesPage.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:33 | Removed unused navigation group import. |
| `app/Filament/Admin/Resources/ApprovalCenter/IpWhitelists/Actions/ApproveIpAction.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:34 | Whitespace and type signature cleanup. |
| `app/Filament/Admin/Resources/ApprovalCenter/IpWhitelists/Actions/EnableIpAction.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:34 | Whitespace and type signature cleanup. |
| `app/Filament/Admin/Resources/ApprovalCenter/IpWhitelists/Actions/RejectIpAction.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:34 | Whitespace and type signature cleanup. |
| `app/Filament/Admin/Resources/ApprovalCenter/IpWhitelists/Actions/RevokeIpAction.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:34 | Whitespace and type signature cleanup. |
| `app/Filament/Admin/Resources/ApprovalCenter/IpWhitelists/Pages/ListIpWhitelistRequests.php` | 2026-09-25 02:08:43 | 2026-10-16 02:23:34 | Removed unused action imports. |
| `app/Filament/Admin/Resources/ApprovalCenter/IpWhitelists/Tables/IpWhitelistRequestsTable.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:35 | Column formatting and table action updates. |
| `app/Filament/Admin/Resources/Finance/Deposits/Tables/DepositsTable.php` | 2026-10-07 10:47:30 | 2026-10-16 02:23:37 | Removed duplicate action handlers. |
| `app/Filament/Admin/Resources/Messaging/Messages/Schemas/MessageInfolist.php` | 2026-09-22 01:34:51 | 2026-10-16 02:23:38 | Removed unused icon imports. |
| `app/Filament/Admin/Resources/Settings/Gateways/Schemas/GatewayForm.php` | 2026-09-22 01:34:52 | 2026-10-16 03:17:16 | Added WhiteInfo gateway credentials fields (`api_token`, `default_sender_id`). |
| `app/Filament/Admin/Resources/Settings/SignupRegisterGateways/Schemas/SignupRegisterGatewayForm.php` | 2026-09-22 01:34:52 | 2026-10-16 01:32:58 | Formatting cleanup. |
| `app/Filament/Admin/Resources/Users/Customers/Actions/ApproveKycAction.php` | 2026-09-22 01:34:52 | 2026-10-16 02:23:41 | Notification phrasing update. |
| `app/Filament/Admin/Resources/Users/Customers/Actions/DepositAction.php` | 2026-10-07 10:07:48 | 2026-10-16 02:23:41 | Transaction reference format update. |
| `app/Filament/Admin/Resources/Users/Customers/Actions/RejectKycAction.php` | 2026-09-22 01:34:52 | 2026-10-16 02:23:41 | Notification phrasing update. |
| `app/Filament/Admin/Resources/Users/Customers/Actions/RequestKycResubmissionAction.php` | 2026-09-22 01:34:52 | 2026-10-16 02:23:42 | Resubmission reason logging update. |
| `app/Filament/Admin/Resources/Users/Customers/Pages/ManageKyc.php` | 2026-09-22 01:34:52 | 2026-10-16 02:23:42 | KYC view layout refinements. |
| `app/Filament/Admin/Resources/Users/Customers/RelationManagers/KycAdminNotesRelationManager.php` | 2026-09-22 01:34:52 | 2026-10-16 02:23:43 | Whitespace cleanup. |
| `app/Filament/Admin/Resources/Users/Customers/Schemas/CustomerInfolist.php` | 2026-10-05 02:05:39 | 2026-10-16 02:23:43 | Infolist section layout refinements. |
| `app/Filament/Admin/Resources/Users/Sps/Actions/DepositAction.php` | 2026-10-07 10:08:10 | 2026-10-16 02:23:43 | Deposit validation and transaction reference update. |
| `app/Filament/Admin/Widgets/AdminTopStatsBarWidget.php` | 2026-09-25 01:27:53 | 2026-10-16 02:23:45 | Metric query cache TTL adjustment. |
| `app/Filament/Admin/Widgets/LiveCustomerActivityWidget.php` | 2026-09-22 01:34:52 | 2026-10-16 02:23:45 | Customer activity polling interval adjustment. |
| `app/Filament/Admin/Widgets/LiveMessageStreamWidget.php` | 2026-09-22 01:34:52 | 2026-10-16 02:23:46 | Live stream query limit adjustment. |
| `app/Filament/Admin/Widgets/LiveQueueStatsWidget.php` | 2026-09-22 01:34:52 | 2026-10-16 02:23:46 | Queue count aggregation tuning. |
| `app/Filament/Admin/Widgets/MissionControlDashboardWidget.php` | 2026-09-23 23:57:26 | 2026-10-16 02:23:47 | Dashboard layout adjustments. |
| `app/Providers/Filament/CustomerPanelProvider.php` | 2026-10-07 00:05:39 | 2026-10-16 02:24:01 | Customer panel plugin registrations cleanup. |
| `app/Providers/Filament/SpPanelProvider.php` | 2026-09-25 03:01:39 | 2026-10-16 03:36:33 | Reseller (SP) panel navigation and route prefix configurations. |
| `app/Providers/Filament/SupermanPanelProvider.php` | 2026-09-25 03:01:28 | 2026-10-16 02:24:01 | Superman panel navigation item ordering. |
| `app/Services/Auth/OtpVerificationService.php` | 2026-10-05 00:02:04 | 2026-10-16 01:32:38 | OTP cooldown and session handling refinements. |
| `app/Services/Sms/SmsManager.php` | 2026-09-22 01:34:55 | 2026-10-16 03:16:58 | Added `createWhiteinfoDriver()` driver resolver method. |
| `app/Settings/GeneralSettings.php` | 2026-09-22 01:34:55 | 2026-10-16 01:32:07 | Settings property definition type annotations. |
| `bootstrap/cache/packages.php` | 2026-10-09 20:11:49 | 2026-10-16 01:56:40 | Re-cached packages including Laravel Horizon and WhiteInfo driver dependencies. |
| `bootstrap/cache/services.php` | 2026-10-09 20:11:49 | 2026-10-16 01:56:40 | Re-cached service providers. |
| `config/app.php` | 2026-09-22 01:34:55 | 2026-10-16 01:31:32 | Provider registration list update. |
| `config/horizon.php` | 2026-09-22 12:03:25 | 2026-10-16 01:58:13 | `SMS-Panel` adds the `'campaigns'` queue supervisor. |
| `config/sms.php` | 2026-09-25 02:40:56 | 2026-10-16 03:15:47 | `SMS-Panel` adds `whiteinfo` gateway configuration block. |
| `database/seeders/GatewaySeeder.php` | 2026-09-22 01:34:55 | 2026-10-16 03:18:12 | `SMS-Panel` adds WhiteInfo gateway seeding record. |
| `database/settings/2026_05_16_131012_create_general_settings.php` | 2026-09-22 01:34:55 | 2026-10-16 01:32:15 | Settings migration formatting. |

### Critical Files Present in `SMS-Panel` (Oct 16) but Missing in Downloads
The following key operational files exist only in `E:\Softcents\SMS-Panel` and do not exist in `C:\Users\asus\Downloads\sms.whiteinfobd.com (1)`:
1. `app/Services/Sms/Drivers/WhiteInfoDriver.php` (created 2026-10-16 03:26:53, 10,043 bytes)
2. `database/migrations/2026_10_16_032000_add_whiteinfo_gateway.php` (created 2026-10-16 03:17:57)
3. `tests/Feature/WhiteInfoDriverTest.php` (created 2026-10-16 03:19:20)
4. `tests/Feature/GatewayFormTest.php` (updated 2026-10-16 03:17:34)

---

## 2. Horizon and Queues

From `config/horizon.php`, `config/queue.php`, `config/database.php`, and `.env`:

### Queue Connection & Prefix Configuration
- **Queue Connection:** `redis` (`QUEUE_CONNECTION=redis`, with supervisor connection set to `'connection' => 'redis'`).
- **Redis Client:** `phpredis` (`REDIS_CLIENT=phpredis`).
- **Database Connection Prefix:**  
  Configured in `config/database.php` as:
  ```php
  'prefix' => env('REDIS_PREFIX', Str::slug((string) env('APP_NAME', 'laravel')).'-database-')
  ```
  With `APP_NAME="SMS-Panel"`, the driver prefix is:
  ```
  sms-panel-database-
  ```
  *Note:* phpredis automatically prepends this connection prefix to every key sent to Redis.
- **Horizon Metadata Prefix:**  
  Configured in `config/horizon.php` as:
  ```php
  'prefix' => env('HORIZON_PREFIX', Str::slug(env('APP_NAME', 'laravel'), '_').'_horizon:')
  ```
  With `APP_NAME="SMS-Panel"`, `HORIZON_PREFIX` evaluates to:
  ```
  sms_panel_horizon:
  ```
  (Stored in raw Redis as `sms-panel-database-sms_panel_horizon:*`).

### Supervisor Configurations

#### 1. Supervisor: `default-1`
- **Queue Handled:** `['default']`
- **Connection:** `redis`
- **Balance Mode:** `auto` (Auto-scaling strategy: `time`)
- **Timeout:** `60` seconds
- **Tries:** `1`
- **Memory Limit:** `128` MB
- **Process Counts:**
  - `defaults`: `maxProcesses` = 1
  - `local`: `maxProcesses` = 3
  - `production`: `maxProcesses` = 10 (`balanceMaxShift` = 1, `balanceCooldown` = 3)

#### 2. Supervisor: `batches`
- **Queue Handled:** `['batches']`
- **Connection:** `redis`
- **Balance Mode:** `auto` (Auto-scaling strategy: `time`)
- **Timeout:** `3600` seconds (1 hour)
- **Tries:** `1`
- **Memory Limit:** `256` MB
- **Process Counts:**
  - `defaults`: `maxProcesses` = 1
  - `local`: `maxProcesses` = 1
  - `production`: `maxProcesses` = 2 (`balanceMaxShift` = 1, `balanceCooldown` = 3)

#### 3. Supervisor: `sms-sending`
- **Queue Handled:** `['sms-sending']`
- **Connection:** `redis`
- **Balance Mode:** `auto` (Auto-scaling strategy: `time`)
- **Timeout:** `3600` seconds (1 hour)
- **Tries:** `1`
- **Memory Limit:** `256` MB
- **Process Counts:**
  - `defaults`: `maxProcesses` = 1
  - `local`: `maxProcesses` = 5
  - `production`: `maxProcesses` = 5 (`balanceMaxShift` = 1, `balanceCooldown` = 3)

#### 4. Supervisor: `campaigns` *(added in Oct 16 snapshot)*
- **Queue Handled:** `['campaigns']`
- **Connection:** `redis`
- **Balance Mode:** `auto` (Auto-scaling strategy: `time`)
- **Timeout:** `3600` seconds (1 hour)
- **Tries:** `1`
- **Memory Limit:** `256` MB
- **Process Counts:**
  - `defaults`: `maxProcesses` = 1
  - `local`: `maxProcesses` = 2
  - `production`: `maxProcesses` = 2 (`balanceMaxShift` = 1, `balanceCooldown` = 3)

### Exact Redis Key Names (for `redis-cli`)

Because the database connection prefix is `sms-panel-database-`, the exact physical keys in Redis are:

#### Queue: `default`
- **Pending Jobs (List):** `sms-panel-database-queues:default`  
  *(Logical key: `queues:default`)*
- **Reserved Jobs (ZSet):** `sms-panel-database-queues:default:reserved`  
  *(Logical key: `queues:default:reserved`)*
- **Delayed Jobs (ZSet):** `sms-panel-database-queues:default:delayed`  
  *(Logical key: `queues:default:delayed`)*

#### Queue: `batches`
- **Pending Jobs (List):** `sms-panel-database-queues:batches`  
  *(Logical key: `queues:batches`)*
- **Reserved Jobs (ZSet):** `sms-panel-database-queues:batches:reserved`  
  *(Logical key: `queues:batches:reserved`)*
- **Delayed Jobs (ZSet):** `sms-panel-database-queues:batches:delayed`  
  *(Logical key: `queues:batches:delayed`)*

#### Queue: `sms-sending`
- **Pending Jobs (List):** `sms-panel-database-queues:sms-sending`  
  *(Logical key: `queues:sms-sending`)*
- **Reserved Jobs (ZSet):** `sms-panel-database-queues:sms-sending:reserved`  
  *(Logical key: `queues:sms-sending:reserved`)*
- **Delayed Jobs (ZSet):** `sms-panel-database-queues:sms-sending:delayed`  
  *(Logical key: `queues:sms-sending:delayed`)*

#### Queue: `campaigns`
- **Pending Jobs (List):** `sms-panel-database-queues:campaigns`  
  *(Logical key: `queues:campaigns`)*
- **Reserved Jobs (ZSet):** `sms-panel-database-queues:campaigns:reserved`  
  *(Logical key: `queues:campaigns:reserved`)*
- **Delayed Jobs (ZSet):** `sms-panel-database-queues:campaigns:delayed`  
  *(Logical key: `queues:campaigns:delayed`)*

---

## 3. Scheduler & Recurring Work

From `routes/console.php` (verified identical in both versions) and `app/Console/`:

### Scheduled Commands
| Command | Frequency | Execution Modifiers | Description |
|---|---|---|---|
| `telescope:prune --hours=48` | `daily()` | `->runInBackground()` | Prunes Telescope entries older than 48 hours. |
| `disposable:update` | `daily()` | `->runInBackground()` | Updates disposable email domains list from remote source. |
| `messages:poll-delivery` | `everyMinute()` | `->withoutOverlapping()->runInBackground()` | Queries gateways for delivery receipts of pending messages. |
| `queue:purge-stale --minutes=8` | `everyMinute()` | `->withoutOverlapping()->runInBackground()` | Marks pending/sending messages & processing batches older than 8 minutes as failed. |
| `batches:send-scheduled` | `everyMinute()` | `->withoutOverlapping()` | Dispatches scheduled message batches whose `scheduled_at <= NOW()`. |
| `campaigns:send-scheduled` | `everyMinute()` | `->withoutOverlapping()` | Dispatches scheduled marketing campaigns whose `scheduled_at <= NOW()`. |

### Other Recurring / Self-Redispatching Work
1. **`PollMessageDeliveryJob` (DLR Backoff Cycle):**  
   When a delivery receipt returns as still submitted/in-transit, `PollMessageDeliveryJob` releases itself back to the queue with incremental backoff delays:  
   **Schedule:** 1m (60s), 2m (120s), 5m (300s), 15m (900s), 1h (3600s), 4h (14400s), 12h (43200s), up to a hard cap of 7 total attempts.
2. **`ProcessMessageBatchJob`:**  
   Splits large recipient files (CSV/XLSX) or contacts into chunks of 1,000 and continuously enqueues `SendBatchBulkSmsJob` / `SendBatchSingleSmsJob` onto the `sms-sending` queue.
3. **`ProcessCampaignJob`:**  
   Resolves target audience contacts and converts them into batches for the `batches` queue.

---

## 4. How It Was Run on the Server

Searched `C:\Users\asus\Downloads\sms.whiteinfobd.com (1)`, the archive `sms.whiteinfobd.com (1).tar.gz`, and `E:\Softcents\SMS-Panel` for:
- Supervisor configs (`*.conf`, `/etc/supervisor/*`)
- Systemd service units (`*.service`)
- Cron lines (`crontab`, `/etc/cron*`)
- Deploy scripts (`deploy.sh`, `run.sh`)
- Nginx server blocks / configs (`nginx.conf`, `*.conf`)

### Result
**NOT IN THIS DOWNLOAD**

*(Note: The only matching filenames found are AI documentation skills under `.agents/skills/configuring-horizon/references/supervisors.md` and an internal lock file `.mimocode/.cron-lock`. No server host configuration files were packaged into the application root archive).*

---

## 5. Storage Layout

### Extraction Status
- **In physical folder `C:\Users\asus\Downloads\sms.whiteinfobd.com (1)`:** `storage/` is **NOT INCLUDED** (not present because extraction halted when Drive C hit 0 bytes free).
- **In `E:\Softcents\SMS-Panel`:** `storage/app/private` contains only `.gitignore`.
- **In the full archive `sms.whiteinfobd.com (1).tar.gz`:** `storage/app/private` is fully intact and contains **250 files**.

### Folder Tree Under `storage/app/private` (from archive)
| Folder Name | File Count | Purpose |
|---|---|---|
| `.` (root of private) | 1 | `.gitignore` |
| `batches` | 53 | Recipient upload files (CSV, XLSX) |
| `filament_exports` | 5 | Admin panel exported spreadsheets |
| `ids` | 73 | KYC ID documents (National ID, Passport scans) |
| `live_photos` | 18 | KYC webcam verification photos |
| `livewire-tmp` | 3 | Temporary uploads from Livewire forms |
| `scribe` | 2 | API documentation artifacts |
| `trade_licenses` | 15 | KYC business trade license documents |
| `uploads` | 10 | General user attachments |

### KYC & Live Photo File Inspection (First Files)

#### 1. First KYC File in `ids/`
- **File Path:** `storage/app/private/ids/01KYF1S0TCK7FV6YCY8GMN5KR8.jpeg`
- **Header Magic Bytes (Hex):** `FF D8 FF E0 00 10 4A 46`
- **Format:** **Plain JPEG image (FF D8)**
- **Encryption:** **NOT encrypted** (stored as standard raw image data).

#### 2. First Live Photo in `live_photos/`
- **File Path:** `storage/app/private/live_photos/live_b3f14a36-5e2e-44ed-8777-cfba67f3a49f.jpg`
- **Header Magic Bytes (Hex):** `FF D8 FF E0 00 10 4A 46`
- **Format:** **Plain JPEG image (FF D8)**
- **Encryption:** **NOT encrypted** (stored as standard raw image data).

*(Note: There is no folder named `kyc/`; customer identity documents are stored in `ids/` and trade licenses in `trade_licenses/`).*

---

## 6. Environment Variable Names

The following variable names appear in `.env` (names only, no secrets or values printed). The Go services must match these names:

### APP_KEY Presence
- **`APP_KEY`**: **PRESENT** in `.env` (standard Laravel base64 key).

### Database Configuration
- `DB_CONNECTION`
- `DB_HOST`
- `DB_PORT`
- `DB_DATABASE`
- `DB_USERNAME`
- `DB_PASSWORD`

### Redis Configuration
- `REDIS_CLIENT`
- `REDIS_HOST`
- `REDIS_PASSWORD`
- `REDIS_PORT`
- `REDIS_DB`
- `REDIS_CACHE_DB`
- `REDIS_PREFIX`

### Queue Configuration
- `QUEUE_CONNECTION`
- `HORIZON_PREFIX`

### Cache Configuration
- `CACHE_STORE`
- `CACHE_PREFIX`
- `MEMCACHED_HOST`

### Mail Configuration
- `MAIL_MAILER`
- `MAIL_SCHEME`
- `MAIL_HOST`
- `MAIL_PORT`
- `MAIL_USERNAME`
- `MAIL_PASSWORD`
- `MAIL_FROM_ADDRESS`
- `MAIL_FROM_NAME`

### Telegram Configuration
- `TELEGRAM_BOT_API_ENABLED`
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_CHAT_ID`

### Payment Provider (PipraPay) Configuration
- `PIPRAPAY_BASE_URL`
- `PIPRAPAY_API_KEY`

### Gateway Provider (WhiteInfo - in Oct 16 Snapshot)
- `WHITEINFO_BASE_URL`
- `WHITEINFO_API_TOKEN`
- `WHITEINFO_DEFAULT_SENDER_ID`

---

## 7. Stale Rows Baseline

Direct SQL query results from the scratch database `sms_panel_scratch` (restored directly from `sms_whiteinfo_2026-10-16.sql`):

### Messages
- **Messages in `status = 'pending'`:** **0**
- **Messages in `status = 'sending'`:** **0**
- **Total Messages in `pending` or `sending`:** **0**
- **Oldest Pending/Sending Message Created Date:** **None (0 rows)**
- *(Total messages in database:* 30, all 30 have `status = 'delivered'`*)*

### Batches
- **Batches in `status = 'pending'`:** **0**
- **Batches in `status = 'processing'`:** **0**
- **Total Batches in `pending` or `processing`:** **0**
- **Oldest Pending/Processing Batch Created Date:** **None (0 rows)**
- *(Total batches in database:* 38, of which 28 have `status = 'completed'` and 10 have `status = 'failed'`*)*

### Queue Tables Baseline
- **Active jobs in `jobs` table:** **0**
- **Failed jobs in `failed_jobs` table:** **10**
