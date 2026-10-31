# SMS Whiteinfo BD — Architecture & Guidelines

> [!CAUTION]
> **CRITICAL: MULTI-TENANT SERVER ISOLATION**  
> This VPS hosts multiple independent production platforms (`account.softcents.com`, `number-portal.softcents.com`, JumpServer).  
> **NEVER** touch, inspect, or modify files outside `/www/wwwroot/smsagri.softcents.com/`.  
> **NEVER** query, modify, or drop any database other than `smsagribd_sc`.  
> **NEVER** execute broad `chmod/chown -R` on `/www` or `/www/wwwroot`.  
> See `SERVER_ISOLATION_RULES.md` for full mandatory isolation guidelines.

## Tech Stack
- **Frontend**: SvelteKit 2 + Svelte 5 (Runes mode) + Tailwind CSS v4 (`frontend/`)
- **Backend**: Go 1.24+ with Go Fiber v3 microservices
  - **Superman Admin API**: `superman-api/` (Port `:8081`, prefix `/api/v1/superman`)
  - **Customer Portal & Developer API**: `customer-api/` (Port `:8082`, prefixes `/api/v1/customer` and `/api/v1/*`)
  - **SP (Reseller) API**: `sp-api/` (Port `:8083`, prefix `/api/v1/sp`)
  - **Master Orchestrator**: `smsagri` (Port `:8080`, manages child microservices)
- **Database**: PostgreSQL 18 (`smsagribd_sc`) via custom driver adapter `internal/pgcompat`
- **Queue / Message Broker**: RabbitMQ (`amqp://127.0.0.1:5672`)
- **No PHP / No Laravel / No Horizon / No Filament**: All backend logic and APIs run natively on Go Fiber.

## Frontend Rules
- Located in `frontend/`.
- Run development server via `npm run dev` inside `frontend/`.
- Run type & Svelte checks via `npm run check`.
- Production build via `npm run build`.
- API client: JWT Bearer token authentication against the Go Fiber backends.
- Never use mock/fake data fallbacks; connect to live Go endpoints.

## Backend Rules
- Written in Go using Go Fiber v3.
- Database connections use `database/sql` with PostgreSQL driver (`github.com/lib/pq`) via `internal/pgcompat` compatibility layer.
- Passwords use bcrypt hashing (cost 12).
- Access tokens use SHA-256 token hashing (compatible with Bearer `<id>|<token>` format).
- Each service has its own `.env` configuration for ports, JWT secrets, PostgreSQL and RabbitMQ credentials.
- Message queuing for SMS dispatch runs via RabbitMQ AMQP; Horizon is strictly deprecated.
