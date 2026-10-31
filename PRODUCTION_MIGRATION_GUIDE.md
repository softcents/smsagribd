# Production Knowledge Transfer & System Handover
## SMS Whiteinfo BD — Migration from PHP Laravel to SvelteKit 2 + Go Fiber Microservices

---

## 1. Executive Summary & Migration Overview

The **SMS Whiteinfo BD** platform has undergone a complete architectural re-engineering, moving from a monolithic PHP Laravel 11/13 + Filament v3/v5 backend to a high-performance, decoupled modern stack:

- **Frontend**: Single-Page Application (SPA) built with **SvelteKit 2**, **Svelte 5 (Runes mode)**, and **Tailwind CSS v4** located in `frontend/`.
- **Backend**: Three lightweight, compiled, independent **Go 1.24+ Fiber v3** microservices (`superman-api`, `customer-api`, `sp-api`).
- **Database**: Single **MySQL** database (`sms_panel`), shared by all services with 100% backwards schema compatibility (bcrypt cost 12 passwords, SHA-256 access tokens).
- **Status of Legacy Code**: All PHP, Laravel, Filament, Blade templates, Composer packages (`vendor/`), and Artisan scripts have been safely decoupled, verified, and completely removed from the workspace. Zero mock, dummy, or fake data remains.

---

## 2. Target Architecture & Port Allocation

```
                     ┌───────────────────────────────────────────────┐
                     │               Nginx Reverse Proxy             │
                     │             (https://sms.whiteinfobd.com)     │
                     └──────┬─────────────────┬───────────────┬──────┘
                            │                 │               │
            Static Files    │                 │               │
     ┌──────────────────────▼───────┐         │               │
     │   SvelteKit Static Build     │         │               │
     │   (frontend/build/index.html)│         │               │
     └──────────────────────────────┘         │               │
                                              │               │
        /api/v1/superman/*                    │               │
        ──────────────────────────────────────┼────────┐      │
        /api/v1/customer/* & /api/v1/*        │        │      │
        ──────────────────────────────────────┼──┐     │      │
        /api/v1/sp/*                          │  │     │      │
        ──────────────────────────────────────┼──┼─────┼──┐   │
                                              │  │     │  │   │
                                    ┌─────────▼┐ │     │  │   │
                                    │ superman │ │     │  │   │
                                    │ -api:8081│ │     │  │   │
                                    └────┬─────┘ │     │  │   │
                                         │  ┌────▼─────┐  │   │
                                         │  │ customer │  │   │
                                         │  │ -api:8082│  │   │
                                         │  └────┬─────┘  │   │
                                         │       │  ┌─────▼─┐ │
                                         │       │  │ sp-api│ │
                                         │       │  │ :8083 │ │
                                         │       │  └───┬───┘ │
                                         │       │      │     │
                                    ┌────▼───────▼──────▼─┐   │
                                    │    MySQL Database   │◄──┘
                                    │     (sms_panel)     │
                                    └─────────────────────┘
```

### Microservices Summary

| Service | Path in Repo | Port | Base Routes | Primary Role |
| :--- | :--- | :--- | :--- | :--- |
| **Superman API** | `superman-api/` | `:8081` | `/api/v1/superman/*` | Admin backoffice: user management, KYC approvals, gateway routing, tariff plans, finances, topbar stats, system settings. |
| **Customer API** | `customer-api/` | `:8082` | `/api/v1/customer/*`<br>`/api/v1/*` | Customer portal endpoints (quick SMS, bulk SMS, contacts, templates, deposits, KYC submission) AND the public External Developer REST API. |
| **SP (Reseller) API** | `sp-api/` | `:8083` | `/api/v1/sp/*` | Service Provider (Reseller) portal: sub-customer list, balance distributions, billing ledger, SMS delivery logs. |
| **Frontend** | `frontend/` | `:5173` (dev)<br>Static (prod) | All UI routes (`/superman/*`, `/customer/*`, `/sp/*`) | SvelteKit static adapter build producing pure HTML, CSS, JS. |

---

## 3. Database & Authentication Compatibility

The Go backend services connect directly to the existing MySQL database (`sms_panel`) using standard `database/sql` and `github.com/go-sql-driver/mysql`.

### 3.1. User & Password Compatibility
- Passwords stored in the `users` table are hashed with standard **bcrypt** (`$2y$` or `$2a$`, cost 12).
- The Go services use `golang.org/x/crypto/bcrypt` for `CompareHashAndPassword` and `GenerateFromPassword(pass, 12)`, making them 100% interoperable with passwords created under Laravel.

### 3.2. Web Portal Authentication (JWT Bearer)
- User logins to `/api/v1/superman/auth/login`, `/api/v1/customer/login`, and `/api/v1/sp/auth/login` authenticate credentials against `users` and issue an **HMAC-SHA256 JWT** Bearer token.
- Default TTL is **8 hours** (or **720 hours / 30 days** if `remember: true`).
- The token payload contains:
  ```json
  {
    "user_id": 1,
    "email": "user@example.com",
    "type": "customer",
    "exp": 1729440000
  }
  ```
- Protected routes use authentication middleware to decode the JWT and set the authenticated context.

### 3.3. External Developer API Authentication (Sanctum Compatible)
- External developers calling `/api/v1/messages`, `/api/v1/me`, `/api/v1/balance`, etc., use tokens generated from the Customer Portal Developer Hub.
- Token format in headers: `Authorization: Bearer <id>|<token>`
- The raw secret `<token>` is hashed using **SHA-256** (`crypto/sha256`) and looked up in the `access_tokens` table:
  ```sql
  SELECT tokenable_id, tokenable_type, abilities 
  FROM access_tokens 
  WHERE id = ? AND token = ?
  ```
- Optional IP Whitelisting is enforced: if a customer has entries in `ip_whitelist_requests` with `status = 'approved'`, requests from unauthorized IPs receive `403 Forbidden`.

### 3.4. Connection Pooling Configuration
All three Go services configure connection pooling in `cmd/server/main.go` to prevent Windows socket aborts and MySQL connection drops:
```go
database.SetMaxOpenConns(25)
database.SetMaxIdleConns(5)
database.SetConnMaxLifetime(5 * time.Minute)
```

---

## 4. Backend Service Breakdown

### 4.1. Superman Admin API (`superman-api/`)
- **Port**: `:8081`
- **Main file**: `cmd/server/main.go`
- **Handlers**: `internal/httpapi/`
  - `auth.go`: Admin login, logout, me, refresh.
  - `topbar.go`: Aggregated statistics for the admin topbar (KYC pending count, total deposits, SMS sent, live low-balance alerts).
  - `dashboard.go`: Revenue charts, delivery statistics, recent activity.
  - `resources.go`: CRUD for customers, resellers (SPs), gateway integrations, tariff plans, operator rates, country rates.
  - `review_actions.go`: Approval center for KYC (`approve`, `reject`, `resubmit`) and IP whitelist requests.
  - `management.go`: User deposit adjustment, account suspension/activation.
  - `system_settings.go`: Platform settings, maintenance mode, global rate caps.

### 4.2. Customer Portal & Developer API (`customer-api/`)
- **Port**: `:8082`
- **Main file**: `cmd/server/main.go`
- **Handlers**: `internal/httpapi/`
  - `auth.go`: Registration, login, password resets, OTP verify/send, user profile, password change.
  - `resources.go`:
    - **SMS Messaging**: Single send (`/messages/send`), scheduled SMS (`/messages/schedule`), bulk send (`/messages/bulk-send`), live queue status.
    - **Batches**: Batch list with pagination, batch detail with per-recipient status breakdown.
    - **Senders / Sender IDs**: Request sender IDs, listing approved sender IDs.
    - **Templates & Audience**: Contact management, group assignments, reusable SMS templates with variable replacement (`{name}`, `{amount}`, etc.).
    - **Finances**: Online/offline deposit requests, transaction history ledger, dynamic operator pricing matrix.
    - **KYC**: Document submission and status polling.
  - `external_api.go`:
    - Replaces Laravel `routes/api.php` for third-party REST API integrations.
    - Endpoints: `GET/POST /api/v1/messages`, `GET /api/v1/me`, `GET /api/v1/balance`, `GET /api/v1/senders`, `GET /api/v1/batches/:batch`, `GET/POST/DELETE /api/v1/ip-whitelists`.
    - Token auth via `access_tokens` table.

### 4.3. SP (Reseller) API (`sp-api/`)
- **Port**: `:8083`
- **Main file**: `cmd/server/main.go`
- **Handlers**: `internal/httpapi/`
  - `auth.go`: SP login, logout, me.
  - `handlers.go`:
    - Reseller dashboard (active balance, assigned sub-customers count, total funds distributed).
    - Sub-customer directory (`/customers`) with filtering.
    - Direct customer balance funding/deduction (`/customers/:id/fund`).
    - Wallet ledger & distribution records (`/wallet`).
    - Reseller billing history (`/billing`).
    - Aggregated SMS logs for all sub-customers (`/messages`).

---

## 5. Frontend Architecture & State Management

The frontend is located entirely in `frontend/` and built with SvelteKit 2 and Svelte 5.

### 5.1. Stores (Svelte 5 Runes)
- `frontend/src/lib/stores/auth.svelte.ts`: Customer session state (`user`, `balance`, `login()`, `logout()`, `updateBalance()`).
- `frontend/src/lib/stores/adminAuth.svelte.ts`: Superman admin session state & real-time topbar metrics.
- `frontend/src/lib/stores/spAuth.svelte.ts`: Reseller session state & wallet balances.
- `frontend/src/lib/stores/toast.svelte.ts`: Global notification toast bus.

### 5.2. API Clients
- `api.ts`: Customer client hitting `/api/v1/customer/*`. Stores token in `localStorage` (`sms_customer_access_token`).
- `superman-api.ts`: Admin client hitting `/api/v1/superman/*`. Stores token in `localStorage` (`sms_superman_access_token`).
- `sp-api.ts`: Reseller client hitting `/api/v1/sp/*`. Stores token in `localStorage` (`sms_sp_access_token`).

### 5.3. Development Proxy (`vite.config.ts`)
```ts
server: {
  port: 5173,
  proxy: {
    '/api/v1/superman': { target: 'http://127.0.0.1:8081', changeOrigin: true },
    '/api/v1/customer': { target: 'http://127.0.0.1:8082', changeOrigin: true },
    '/api/v1/sp':       { target: 'http://127.0.0.1:8083', changeOrigin: true },
    '/api/v1':          { target: 'http://127.0.0.1:8082', changeOrigin: true }
  }
}
```

---

## 6. Production Deployment Guide (Step-by-Step for Live Server)

This guide is designed for deploying to a standard Linux server (Ubuntu 22.04 / 24.04 LTS or Debian 12) with Nginx and MySQL.

### 6.1. Server Prerequisites
Install Go 1.24+, Node.js 20+ LTS, and Nginx:
```bash
# Update packages
sudo apt update && sudo apt install -y curl git nginx build-essential

# Install Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# Install Go 1.24
wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 6.2. Directory Layout on Live Server
Suggested production folder structure under `/var/www/sms.whiteinfobd.com`:
```
/var/www/sms.whiteinfobd.com/
├── frontend/
│   ├── build/              <-- Built static site files (HTML/CSS/JS)
│   └── package.json
├── superman-api/
│   ├── superman-api        <-- Compiled Linux binary
│   └── .env
├── customer-api/
│   ├── customer-api        <-- Compiled Linux binary
│   └── .env
└── sp-api/
    ├── sp-api              <-- Compiled Linux binary
    └── .env
```

### 6.3. Compiling Go Binaries for Linux
If building on your Windows development machine for Linux (cross-compilation):
```powershell
# In PowerShell (root of repository)
$env:GOOS = "linux"
$env:GOARCH = "amd64"

# Build superman-api
cd superman-api
go build -ldflags="-s -w" -o superman-api ./cmd/server

# Build customer-api
cd ../customer-api
go build -ldflags="-s -w" -o customer-api ./cmd/server

# Build sp-api
cd ../sp-api
go build -ldflags="-s -w" -o sp-api ./cmd/server

# Reset env
$env:GOOS = "windows"
```
Or if compiling directly on the Linux server:
```bash
cd /var/www/sms.whiteinfobd.com/superman-api && go build -ldflags="-s -w" -o superman-api ./cmd/server
cd /var/www/sms.whiteinfobd.com/customer-api && go build -ldflags="-s -w" -o customer-api ./cmd/server
cd /var/www/sms.whiteinfobd.com/sp-api && go build -ldflags="-s -w" -o sp-api ./cmd/server

# Make binaries executable
chmod +x /var/www/sms.whiteinfobd.com/*-api/*-api
```

### 6.4. Building the Frontend Production Assets
```bash
cd /var/www/sms.whiteinfobd.com/frontend
npm ci
npm run build
# Outputs the static production site into frontend/build/
```

### 6.5. Environment Configuration Files

#### `/var/www/sms.whiteinfobd.com/superman-api/.env`:
```ini
SUPERMAN_API_ADDR=:8081
SUPERMAN_API_ENV=production
SUPERMAN_API_JWT_SECRET=7nWq3Dx2Vu8Lr5Ks9Aa4Hj6Mp1Zc0YtQe3Bi7Fo2Gl8Rx5Nd
SUPERMAN_API_FRONTEND_ORIGIN=https://sms.whiteinfobd.com
SUPERMAN_API_TOKEN_TTL=8h
SUPERMAN_API_REMEMBER_TOKEN_TTL=720h

DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=sms_panel
DB_USERNAME=root
DB_PASSWORD=YOUR_PROD_DB_PASSWORD
```

#### `/var/www/sms.whiteinfobd.com/customer-api/.env`:
```ini
CUSTOMER_API_ADDR=:8082
CUSTOMER_API_ENV=production
CUSTOMER_API_JWT_SECRET=kC9mP2nR7vX4wL8qA5hJ3bY0sD6tE1uZ7fN2gM8oK
CUSTOMER_API_FRONTEND_ORIGIN=https://sms.whiteinfobd.com
CUSTOMER_API_TOKEN_TTL=8h
CUSTOMER_API_REMEMBER_TOKEN_TTL=720h

DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=sms_panel
DB_USERNAME=root
DB_PASSWORD=YOUR_PROD_DB_PASSWORD
```

#### `/var/www/sms.whiteinfobd.com/sp-api/.env`:
```ini
SP_API_ADDR=:8083
SP_API_ENV=production
SP_API_JWT_SECRET=Sp4KxN2Wq7Yr1Tz9Vm5Lk3Jb6Hu8Ao2Dp0Ec4Fg7Is1Mn3Pw
SP_API_FRONTEND_ORIGIN=https://sms.whiteinfobd.com
SP_API_TOKEN_TTL=8h
SP_API_REMEMBER_TOKEN_TTL=720h

DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=sms_panel
DB_USERNAME=root
DB_PASSWORD=YOUR_PROD_DB_PASSWORD
```

---

### 6.6. Systemd Service Units

Create 3 systemd services to ensure the Go microservices run in the background, start on system boot, and auto-restart on crashes.

#### 1. Superman Admin Service: `/etc/systemd/system/superman-api.service`
```ini
[Unit]
Description=SMS Whiteinfo Superman Admin API (Go Fiber)
After=network.target mysql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/var/www/sms.whiteinfobd.com/superman-api
ExecStart=/var/www/sms.whiteinfobd.com/superman-api/superman-api
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

#### 2. Customer API Service: `/etc/systemd/system/customer-api.service`
```ini
[Unit]
Description=SMS Whiteinfo Customer & Developer API (Go Fiber)
After=network.target mysql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/var/www/sms.whiteinfobd.com/customer-api
ExecStart=/var/www/sms.whiteinfobd.com/customer-api/customer-api
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

#### 3. SP Reseller Service: `/etc/systemd/system/sp-api.service`
```ini
[Unit]
Description=SMS Whiteinfo SP Reseller API (Go Fiber)
After=network.target mysql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/var/www/sms.whiteinfobd.com/sp-api
ExecStart=/var/www/sms.whiteinfobd.com/sp-api/sp-api
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

#### Enable and Start Services:
```bash
sudo systemctl daemon-reload
sudo systemctl enable superman-api customer-api sp-api
sudo systemctl start superman-api customer-api sp-api

# Verify status
sudo systemctl status superman-api customer-api sp-api
```

---

### 6.7. Complete Nginx Server Configuration

Save this file as `/etc/nginx/sites-available/sms.whiteinfobd.com`:

```nginx
server {
    listen 80;
    server_name sms.whiteinfobd.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name sms.whiteinfobd.com;

    # SSL Certificates (managed via Certbot or custom)
    ssl_certificate /etc/letsencrypt/live/sms.whiteinfobd.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sms.whiteinfobd.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Root directory pointing to built static frontend
    root /var/www/sms.whiteinfobd.com/frontend/build;
    index index.html;

    # Max upload limit for KYC files & bulk CSV uploads
    client_max_body_size 25M;

    # Gzip Compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript image/svg+xml;

    # 1. Superman Admin API reverse proxy
    location /api/v1/superman/ {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 2. Customer Portal API reverse proxy
    location /api/v1/customer/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 3. SP (Reseller) API reverse proxy
    location /api/v1/sp/ {
        proxy_pass http://127.0.0.1:8083;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 4. External Developer REST API reverse proxy (/api/v1/messages, /api/v1/balance, etc.)
    location /api/v1/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 5. Static Assets Caching
    location ~* \.(?:css|js|jpg|jpeg|gif|png|ico|cur|gz|svg|svgz|mp4|ogg|ogv|webm|htc|woff2|woff)$ {
        expires 1y;
        access_log off;
        add_header Cache-Control "public, immutable";
        try_files $uri =404;
    }

    # 6. SvelteKit SPA fallback routing
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

Enable site and test Nginx:
```bash
sudo ln -s /etc/nginx/sites-available/sms.whiteinfobd.com /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## 7. Verification & Sanity Check Procedures

Once deployed, verify all services:

1. **Test Go API Health Endpoints**:
   ```bash
   curl -s http://127.0.0.1:8081/health
   # Expected: {"status":"ok"}

   curl -s http://127.0.0.1:8082/health
   # Expected: {"status":"ok"}

   curl -s http://127.0.0.1:8083/health
   # Expected: {"status":"ok"}
   ```

2. **Test External Developer API Auth via Domain**:
   ```bash
   curl -i https://sms.whiteinfobd.com/api/v1/me
   # Expected: HTTP/2 401 Unauthorized with {"message":"Unauthenticated."}
   ```

3. **Test Customer Portal Login**:
   ```bash
   curl -i -X POST https://sms.whiteinfobd.com/api/v1/customer/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"invalidpassword"}'
   # Expected: HTTP 401 Unauthorized with {"message":"Invalid credentials."}
   ```

4. **Verify Frontend SPA Delivery**:
   Visit `https://sms.whiteinfobd.com` in a browser. The login screen loads instantly without SSR dependencies or PHP overhead.
