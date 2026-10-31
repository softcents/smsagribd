# SMS Aggregator Panel

High-performance Telecom SMS Gateway Aggregator and Management Platform built with **Go (Golang)** microservices, **PostgreSQL**, **RabbitMQ**, and a modern **SvelteKit 5** single-page application.

---

> [!CAUTION]
> ### 🛡️ Multi-Tenant Server Isolation Directive
> This server hosts multiple independent live platforms managed by aaPanel (e.g. `account.softcents.com`, `number-portal.softcents.com`, JumpServer).
> - **Directory Scope**: Operations must remain strictly within `/www/wwwroot/smsagri.softcents.com/`. Never touch or modify sibling directories.
> - **Database Scope**: Operations must remain strictly within database `smsagribd_sc` (user: `smsagribd_sc`). Never touch `ac_sc` or `noportal_sc`.
> - **Port Scope**: Bindings are strictly limited to `127.0.0.1:8080-8083`.
> - See [`SERVER_ISOLATION_RULES.md`](./SERVER_ISOLATION_RULES.md) for full details.

---

## Architecture & Technology Stack

- **Architecture**: Microservices orchestrated via single master process (`smsagri`)
- **Backend**: Go 1.24+ using Go Fiber v3
  - **Superman Admin API**: `superman-api/` (Port `:8081`, prefix `/api/v1/superman`)
  - **Customer Portal & Developer API**: `customer-api/` (Port `:8082`, prefixes `/api/v1/customer`, `/api/v1/`)
  - **SP (Reseller) API**: `sp-api/` (Port `:8083`, prefix `/api/v1/sp`)
  - **Master Orchestrator / Status Monitor**: `main.go` -> `smsagri` (Port `:8080`)
- **Frontend**: SvelteKit 2 + Svelte 5 (Runes mode) + Tailwind CSS v4 (`frontend/`)
- **Database**: PostgreSQL 18 (`smsagribd_sc`) via custom driver adapter `internal/pgcompat`
- **Queue / Broker**: RabbitMQ (AMQP `:5672`, Management UI `:15672`)
- **Web Server / Reverse Proxy**: Nginx + SSL
- **Process Manager**: aaPanel Go Project Manager / Systemd

---

## Directory Structure

```text
├── build.sh                  # One-click build script for all Go binaries
├── git-auto-push.sh          # Automated Git push & sync script
├── git-auto-pull.sh          # Automated Git pull, rebuild, and reload script
├── main.go                   # Master Orchestrator daemon source
├── customer-api/             # Customer SMS Engine, Queues, Billing & API
├── superman-api/             # Administrator & Platform Operations API
├── sp-api/                   # Reseller / SP Management API
├── frontend/                 # SvelteKit 5 Source & Static SPA Build
├── database/                 # PostgreSQL schema and migration dumps
│   ├── schema_pgsql.sql
│   ├── dump_pgsql.sql
│   └── setup.sh
├── deploy/                   # Deployment guides and env templates
├── public/                   # Public static branding assets & storage
├── SERVER_ISOLATION_RULES.md # Critical server isolation rules
└── storage/                  # KYC documents, uploads, and batch CSVs
```

---

## Quick Start & Setup

### 1. Database Setup
```bash
./database/setup.sh
```

### 2. Build Go Microservices
```bash
./build.sh
```

### 3. Run Microservices
```bash
./smsagri
```

### 4. Verification Check
```bash
./customer-api/selfcheck
```

---

## License
Proprietary software © Softcents. All rights reserved.
