# Arkive Architecture Notes (Core + Cloud)

---

# 🧠 Core Philosophy (CRITICAL)

Arkive is built as **two layers**:

## 🟢 Arkive Core (Open Source)

> A self-hostable, single-user, zero-knowledge file sharing server

## 🔵 Arkive Cloud (SaaS)

> A hosted, scalable, multi-user, high-performance version of Arkive

---

## Design Principles

* **Trust-first** → encryption + storage logic must be open
* **Separation of concerns** → Core = system, Cloud = experience
* **Single-user first (Core)** → no multi-user complexity in OSS
* **Performance & scale (Cloud)** → optimized infra lives here
* **Config over plans (Core)** → no pricing logic in OSS
* **Safe defaults** → limits exist but are configurable

---

# 🏗️ Architecture Overview

```text id="arch1"
arkive/
├── core/        # 🟢 Open source (Go backend + minimal UI)
├── crypto/      # 🟢 Open source (client-side encryption)
├── web/         # 🟡 Client (shared, may diverge)
├── cloud/       # 🔵 Private SaaS layer
```

---

# 🟢 Arkive Core Responsibilities

## Auth

* Single-user system
* Email/password login
* No Google OAuth
* Email verification optional (disabled by default)

---

## Storage

* Local disk storage (default)
* S3-compatible (R2, Backblaze, Wasabi, AWS)
* Configurable via setup or env

---

## Files

* Upload (simple)
* Download
* Delete
* List (basic pagination)

---

## Sharing

* Public share links
* Password protection (optional)
* Expiry (per file/share)

---

## Setup & Onboarding

* `/setup` wizard
* Instance initialization flag
* Storage selection (local or S3)
* Admin account creation

---

## Limits (Config-based ONLY)

```env id="core_limits"
MAX_FILE_SIZE
MAX_STORAGE_BYTES
MAX_UPLOAD_CONCURRENCY
UPLOAD_STALE_HOURS
RATE_LIMIT_RPM
```

Rules:

* 0 = unlimited
* No pricing tiers
* No plan-based logic

---

## Background Jobs

### Keep:

* cleanup_cron.go → removes stale uploads
* expiry_cron.go → deletes expired files

### Remove:

* retention_cron.go (inactive user deletion)

---

## Rate Limiting

* Basic per-IP/global limiter
* Config-based
* No dynamic scaling

---

## Mailer (Abstracted)

```go id="mailer_core"
type Mailer interface {
    Send(msg Message) error
}
```

Implementations:

* SMTP (default)
* Noop (dev)

Cloud-only:

* Postmark adapter

---

## Web UI (Core)

Minimal only:

* Upload
* File list
* Share links
* Delete
* Storage usage

Removed:

* Landing page
* Pricing
* SEO pages
* Ads

---

## Routing (Core)

### Public

* `/setup`
* `/login`
* `/share/:token`

### Protected

* `/dashboard`
* `/settings`

### Root (`/`)

* Not initialized → `/setup`
* Not authenticated → `/login`
* Authenticated → `/dashboard`

---

# 🔵 Arkive Cloud Responsibilities

## Auth & Users

* Multi-user system
* OAuth (Google, etc.)
* Email verification (required)
* Account recovery flows

---

## Storage & Performance

* Multipart uploads
* Resumable uploads
* Parallel chunking
* CDN-backed delivery

---

## Sharing (Advanced)

* Link revoke/update
* Download limits
* Advanced permissions
* Share analytics

---

## Analytics & Insights

* File views/downloads
* Bandwidth usage
* Activity tracking

---

## Rate Limiting (Advanced)

* Per-user limits
* Plan-based scaling
* Abuse detection

---

## Retention Policies

* Inactive account cleanup
* Storage tiering
* Automated archiving

---

## Mailer (Production)

* Postmark integration
* Retry logic
* Delivery tracking
* Templates

---

## UI / UX

* Polished dashboard
* Previews (image/video)
* Drag-drop uploads
* Growth pages

---

## Growth & Marketing

* Landing pages
* Pricing
* Contact
* SEO
* Legal pages

---

## Billing

* Plans
* Quotas
* Payment integration

---

# 🧩 Shared Components

## Crypto (Open Source)

* Client-side encryption
* KEK/DEK model
* Key wrapping
* Share keys

## API Contract

* Shared between Core and Cloud
* Platform-agnostic (web, mobile, CLI)

---

# ⚖️ Separation Rules (IMPORTANT)

| Rule                | Explanation               |
| ------------------- | ------------------------- |
| Trust → Core        | encryption, storage logic |
| Performance → Cloud | speed, CDN, scaling       |
| UX polish → Cloud   | smooth experience         |
| Config → Core       | limits, storage           |
| Plans → Cloud       | pricing, quotas           |

---

# 🔐 Security Model

* Client-side encryption (zero-knowledge)
* Server stores encrypted blobs only
* HTTPS handled via proxy (Tailscale, Cloudflare Tunnel)

---

# 🧠 Conventions

* SQL only in repositories
* Services own transactions
* The server only stores encrypted file and folder metadata. Pages may render encrypted blobs, but plaintext names and details must be decrypted client-side through the unlocked user vault.
* Handlers remain thin
* Validation via `pkg/validation`
* Errors centralized in services
* Minimal dependencies
* No SaaS logic in Core
* Keep SQL projections tight and only select columns the caller actually needs
* Don't add unnecessary code or helper functions; only extract helpers when duplication is substantial or the abstraction is clearly necessary
* In JavaScript, prefer `async`/`await` where practical instead of promise chains
* In markup for pages etc, create components.go and put helper markup functions there so the main file is not cluttered
---

# ⚠️ Edge Cases

* No users → system resets to `/setup`
* Setup interrupted → restart allowed
* Missing config → fallback defaults

---

# 🚀 Future Roadmap

## Core

* Multi-user (optional future)
* CLI client
* Better storage adapters

## Cloud

* Mobile apps
* Sync engine
* Enterprise features
* Advanced sharing

---

# 🔥 Key Philosophy (Do NOT break)

## Core

> Works fully, but simple and manual

## Cloud

> Fast, scalable, and polished

---

# 🧠 Final Guideline

> If removing a feature breaks self-hosting → keep in Core
> If removing a feature only reduces convenience → move to Cloud

---

# Don't Forget

* Keep Core simple and predictable
* Avoid vendor lock-in in Core
* Do not auto-delete user data without explicit rules
* Prioritize trust over features
* Maintain strict separation between Core and Cloud
