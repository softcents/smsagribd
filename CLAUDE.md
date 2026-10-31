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
- **Database**: PostgreSQL 18 (`smsagribd_sc`)
- **Queue**: RabbitMQ (AMQP 5672)
- **No PHP / No Laravel / No Horizon / No Filament**: All backend logic and APIs run natively on Go Fiber.
