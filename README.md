# Arkive Core

> Open-source, self-hostable, zero-knowledge encrypted file storage and sharing.

Arkive Core is a privacy-focused file storage platform built around a simple idea:

**Your files should be encrypted before they leave your device.**

Files and sensitive metadata are encrypted client-side. The server stores encrypted blobs and encrypted metadata only. File names, folder names, manifests, thumbnails, and file contents are decrypted client-side after the vault is unlocked.

Arkive Core is designed to provide modern cloud storage functionality while preserving privacy through a zero-knowledge architecture.

---

## Features

### Zero-Knowledge Encryption

- Client-side file encryption
- Client-side metadata encryption
- Client-side folder encryption
- Vault unlock, lock, and re-unlock flow
- Recovery keys
- Password reset and master-key rewrapping
- Zero-knowledge share links
- Privacy-preserving search

### Storage

- Local filesystem storage
- S3-compatible storage providers
- Cloudflare R2 support
- Multipart uploads
- Direct-to-storage uploads via presigned URLs
- Large file support

### File Management

- Files and folders
- Drag-and-drop uploads
- Move, rename, and delete
- Folder hierarchy
- Grid and list views
- Encrypted search

### Sharing

- Public share links
- Password-protected shares
- Expiring links
- One-time links (burn-after-read)
- Share revocation and restoration

---

## What Makes Arkive Different?

Arkive is built around a simple rule:

> The server should never need access to your data.

| Data | Server Can Read |
|--------|--------|
| File contents | ❌ |
| File names | ❌ |
| Folder names | ❌ |
| File metadata | ❌ |
| Encryption keys | ❌ |
| Upload status | ✅ |
| Storage usage | ✅ |
| Account email | ✅ |

The backend only stores encrypted data and encrypted metadata. Decryption happens entirely on the client after the vault is unlocked.

---

## Architecture

Arkive Core is a server-rendered Go application with client-side cryptography powered by Rust and WebAssembly.

```text
Browser
 ├─ Vault unlock
 ├─ Encrypt files
 ├─ Encrypt metadata
 ├─ Generate search tokens
 └─ Decrypt content
          │
          ▼
Arkive Core API (Go)
          │
          ▼
Storage Provider
(Local / S3-compatible / R2)
```

All encryption and decryption happens on the client after vault unlock.

The backend never receives plaintext file names, folder names, manifests, thumbnails, or file contents.

---

## Privacy-Preserving Search

Arkive supports encrypted file and folder search.

Search tokens are derived client-side from the unlocked vault and sent to the server as blind lookup tokens. The server never receives plaintext search terms.

Like most searchable-encryption systems, token equality patterns may be observable, but plaintext names and search queries remain hidden.

---

## Cryptography

Client-side cryptography is implemented using the Arkive Crypto WebAssembly library.

Algorithms currently used include:

- XChaCha20-Poly1305
- Argon2id
- HKDF-SHA256
- BLAKE3
- SHA-256

Arkive Crypto is available separately and contains implementation details, encryption formats, key derivation logic, and cryptographic documentation.

See:

- Arkive Crypto repository
- `core/web/static/vendor/arkive-crypto/README.md`

---

## Ecosystem

Arkive consists of multiple projects that work together:

- Arkive Core — storage server and web application
- Arkive Crypto — Rust/WASM cryptography library
- Arkive Sync — native synchronization engine (planned)
- Arkive Desktop — desktop application (planned)
- Arkive Mobile — mobile applications (planned)

All projects share the same zero-knowledge architecture and client-side cryptography model.

---

## Run Locally

Use this for development on your machine.

### Requirements

- Go
- Docker
- Docker Compose

### Clone

```bash
git clone https://github.com/numanjvd82/arkive.git
cd arkive
```

### Configure

```bash
cp .env.example .env
```

For local dev, keep these values:

- `APP_ENV=dev`
- `DATABASE_URL=postgres://devuser:devpassword123!@localhost:5132/arkive_dev?sslmode=disable`
- `COOKIE_SECRET`
- `SESSION_TTL`

### Start PostgreSQL

```bash
docker compose up -d db
```

### Start app

```bash
make dev
```

Open:

```text
http://localhost:8080
```

---

## Deploy on a VPS

Use this path when you want Arkive public on a server.

### 1. Provision server

- Linux VPS
- Docker installed
- Docker Compose installed
- Domain name pointed at VPS IP

### 2. Get code

```bash
git clone https://github.com/numanjvd82/arkive.git
cd arkive
```

### 3. Configure production env

```bash
cp .env.example .env
```

Set these values in `.env`:

- `APP_ENV=prod`
- `BASE_URL=https://your-domain.com`
- `COOKIE_SECRET` to strong random secret
- `SESSION_TTL` to desired duration

Keep PostgreSQL private. App container uses `db:5432` over Docker network.

### 4. Build and start

```bash
docker compose up -d --build
```

### 5. Open app

Visit:

```text
http://your-server-ip:8080
```

If you want domain + HTTPS, put a reverse proxy in front of port `8080` and point `BASE_URL` to the public HTTPS URL.

### 6. Update

```bash
git pull
docker compose up -d --build
```

---

## Screenshots

Screenshots and demos will be added as the project evolves.

---

## Project Goals

Arkive aims to provide a modern, privacy-focused storage platform built around zero-knowledge principles.

The project is designed to be:

- Open source
- Self-hostable
- Privacy-first
- Storage-provider agnostic
- Extensible for future clients and services

Future development may include desktop applications, mobile applications, sync clients, and additional deployment models while preserving the project's zero-knowledge architecture.

---

## Security

Arkive is designed around a zero-knowledge architecture.

However:

- Self-hosted administrators control the infrastructure they deploy.
- Browser security remains important.
- Losing recovery material may permanently lock encrypted data.

Arkive has not yet completed a third-party security audit.

Users should independently evaluate whether Arkive meets their threat model before storing sensitive information.

---

## Contributing

Contributions are welcome.

Important architectural rules:

- SQL belongs in repositories.
- Services own business logic and transactions.
- Handlers stay thin.
- Zero-knowledge guarantees must be preserved.
- Plaintext file and folder names must never be visible to the backend.
- Storage object names must never be changed during moves or renames.

---

## Roadmap

### Core

- [x] Client-side encryption
- [x] Multipart uploads
- [x] Files and folders
- [x] Sharing
- [x] Search
- [x] Vault lock and re-unlock flow

### Planned

- [ ] Native sync engine
- [ ] Desktop applications
- [ ] Mobile applications
- [ ] Additional storage providers
- [ ] File versioning
- [ ] Offline support

---

## License

AGPL-3.0

See [LICENSE](./LICENSE) for details.
