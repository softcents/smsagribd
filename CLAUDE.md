# SMS Whiteinfo BD — Architecture & Guidelines

## Tech Stack
- **Frontend**: SvelteKit 2 + Svelte 5 (Runes mode) + Tailwind CSS v4 (`frontend/`)
- **Backend**: Go 1.24+ with Go Fiber v3 microservices
  - **Superman Admin API**: `superman-api/` (Port `:8081`, prefix `/api/v1/superman`)
  - **Customer Portal & Developer API**: `customer-api/` (Port `:8082`, prefixes `/api/v1/customer` and `/api/v1/*`)
  - **SP (Reseller) API**: `sp-api/` (Port `:8083`, prefix `/api/v1/sp`)
- **Database**: MySQL (`sms_panel`)
- **No PHP / No Laravel / No Filament**: All backend logic and APIs run natively on Go Fiber.
