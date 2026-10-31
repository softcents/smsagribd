# SMS Whiteinfo BD — Architecture & Guidelines

## Tech Stack
- **Frontend**: SvelteKit 2 + Svelte 5 (Runes mode) + Tailwind CSS v4 (`frontend/`)
- **Backend**: Go 1.24+ with Go Fiber v3 microservices
  - **Superman Admin API**: `superman-api/` (Port `:8081`, prefix `/api/v1/superman`)
  - **Customer Portal & Developer API**: `customer-api/` (Port `:8082`, prefixes `/api/v1/customer` and `/api/v1/*`)
  - **SP (Reseller) API**: `sp-api/` (Port `:8083`, prefix `/api/v1/sp`)
- **Database**: MySQL (`sms_panel`)
- **No PHP / No Laravel / No Filament**: All backend logic and APIs run natively on Go Fiber.

## Frontend Rules
- Located in `frontend/`.
- Run development server via `npm run dev` inside `frontend/`.
- Run type & Svelte checks via `npm run check`.
- Production build via `npm run build`.
- API client: JWT Bearer token authentication against the Go Fiber backends.
- Never use mock/fake data fallbacks; connect to live Go endpoints.

## Backend Rules
- Written in Go using Go Fiber v3.
- Database connections use `database/sql` with MySQL driver and configured connection pool (`SetConnMaxLifetime`, `SetMaxOpenConns`, `SetMaxIdleConns`).
- Passwords use bcrypt hashing (compatible with standard 60-character cost 12 bcrypt hashes).
- Access tokens use SHA-256 token hashing (compatible with Bearer `<id>|<token>` format).
- Each service has its own `.env` configuration for ports, JWT secrets, and MySQL credentials.
