# SERVER ISOLATION & MULTI-TENANCY GUIDELINES
> **CRITICAL DIRECTIVE FOR ALL DEVELOPERS & AI ASSISTANTS**  
> This server runs **multiple production platforms simultaneously** managed via aaPanel. Any maintenance, development, or deployment on `smsagri.softcents.com` must **NEVER** interfere with, modify, or impact any other application on this machine.

---

## 1. Server Environment Landscape

This VPS hosts multiple active production platforms:

| Application / Site | Directory | Associated Databases / Services |
|---|---|---|
| **Identity & Account Platform** | `/www/wwwroot/account.softcents.com` | `ac_sc` (PostgreSQL), Next.js / NestJS |
| **Number Portal Platform** | `/www/wwwroot/number-portal.softcents.com` | `noportal_sc` (PostgreSQL) |
| **JumpServer** | `/www/server/panel/plugin/jumpserver` | JumpServer core services |
| **SMS Aggregator (This Project)** | `/www/wwwroot/smsagri.softcents.com` | `smsagribd_sc` (PostgreSQL), RabbitMQ, Go Fiber |

---

## 2. Strict Isolation Rules

### Rule 1: Directory Scope Lock
* **ONLY modify files within `/www/wwwroot/smsagri.softcents.com/`.**
* **STRICTLY PROHIBITED:** Never edit, move, delete, or inspect directories belonging to other applications (e.g. `/www/wwwroot/account.softcents.com`, `/www/wwwroot/number-portal.softcents.com`).
* **STRICTLY PROHIBITED:** Never execute broad recursive permission commands such as `chown -R` or `chmod -R` on parent directories (`/www`, `/www/wwwroot`, `/var`, `/etc`). Always explicitly target `/www/wwwroot/smsagri.softcents.com/`.

### Rule 2: Database Scope Lock
* **Target Database:** `smsagribd_sc`
* **Target User:** `smsagribd_sc`
* **STRICTLY PROHIBITED:** Never drop, truncate, alter, connect to, or query any other database (such as `ac_sc`, `noportal_sc`, or system tables).
* All connection strings and setup scripts must explicitly scope connections to `smsagribd_sc`.

### Rule 3: Network & Port Allocation Lock
* This project is strictly allocated the following internal loopback ports:
  - **Master Orchestrator:** `127.0.0.1:8080`
  - **Superman Admin API:** `127.0.0.1:8081`
  - **Customer Portal API:** `127.0.0.1:8082`
  - **Service Provider (SP) API:** `127.0.0.1:8083`
* **STRICTLY PROHIBITED:** Never bind services to non-allocated ports or `0.0.0.0` without verification, to prevent port collisions with other applications.

### Rule 4: Shared Daemons & Service Management
* PostgreSQL, RabbitMQ, and Nginx are **shared system infrastructure**.
* Never blindly restart or stop shared daemons (`systemctl restart postgresql`, `systemctl stop rabbitmq-server`).
* When modifying Nginx configuration, **ONLY** edit `/www/server/panel/vhost/nginx/go_smsagri_softcents_com.conf`. Always test with `nginx -t` before reloading.
* Never kill processes by broad names like `killall -9 node` or `killall -9 python` as this will kill other applications running on the server.

### Rule 5: aaPanel Configuration Safety
* The server uses aaPanel (`/www/server/panel`).
* Never manually corrupt or directly lock `/www/server/panel/data/default.db`.
* Cron jobs, certificates, and site configs must be managed following aaPanel conventions so that all applications remain fully visible and manageable in the aaPanel UI.
