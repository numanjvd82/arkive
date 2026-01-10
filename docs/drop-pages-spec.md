# Drop Pages (Collections) Spec

## Goal
Create a stable, shareable page that can host multiple files and be updated over time.
This turns Arkive from one-off sharing into a maintained, revisited destination.

Drop Pages can be used to share more than files, as long as content follows Arkive guidelines:
- Code samples, demos, and project builds.
- Image collections and video reels.
- Portfolios, media kits, and presentations.
- Assignments, study packs, and slideshows.

## Naming
- Internal: Collections
- User-facing: Drop Pages (or Share Pages)

## URL Shape
Choose one primary shape and keep it stable:
- Preferred: `arkive.sh/@{username}/{slug}`
- Alternate: `arkive.sh/p/{slug}`

## MVP Scope (v1)
- Create page with title and slug.
- Add files to page.
- Reorder files.
- Optional password.
- Share stable link.
- View count + download count.

## User Stories
- As a creator, I make a media kit page and update it without changing the link.
- As a freelancer, I deliver files to a client with a branded page.
- As a student, I share a course pack and update it after lectures.
- As a viewer, I can browse and download without logging in.
- As a developer, I share code samples, demos, and build artifacts.
- As a designer, I publish a portfolio with images and videos.

## UX Flows

### Create Page
1) User opens "New Drop Page".
2) Inputs: title, slug, description (optional), cover image (optional).
3) Privacy: public / unlisted / password / expiry (optional for MVP).
4) Create -> redirect to manage view.

### Manage Page
- Summary header: title, link, visibility, last updated.
- File picker to attach existing uploads.
- Drag to reorder.
- Per-file actions: remove from page (does not delete file).
- Stats: page views, file downloads.

### Public Page
- Hero: title, description, cover.
- List/grid of files with type badges and file size.
- Primary action: Download or View (if previewable).
- Optional password gate.
- Optional use-case highlights (creator, developer, student, team).

## Data Model

### drop_pages
- id (uuid pk)
- owner_id (uuid fk users)
- title (text)
- slug (text, unique per owner)
- description (text, nullable)
- cover_file_id (uuid fk files, nullable)
- visibility (enum: public, unlisted, password)
- password_hash (text, nullable)
- expires_at (timestamp, nullable)
- created_at (timestamp)
- updated_at (timestamp)
- published_at (timestamp, nullable)
- view_count (bigint, default 0)

### drop_page_items
- id (uuid pk)
- page_id (uuid fk drop_pages)
- file_id (uuid fk files)
- position (int)
- created_at (timestamp)

### drop_page_views
Optional for analytics details (can be added later):
- id (uuid pk)
- page_id (uuid)
- viewer_hash (text) // ip+ua hash
- created_at (timestamp)

### drop_page_downloads
Optional for analytics details:
- id (uuid pk)
- page_id (uuid)
- file_id (uuid)
- viewer_hash (text)
- created_at (timestamp)

## Permissions
- Only owner can create/manage/drop pages.
- Public/unlisted view: no login required.
- Password: prompt before revealing list.
- Expiry: returns 410 or "expired" landing page.
- Remove file from page: does not delete file.

## Repo and Service Layout
Follow existing layering:
- Handlers: parse inputs and call service.
- Services: validation, tx, orchestration.
- Repos: SQL only.

Suggested packages:
- `core/handlers/drop_pages.go`
- `core/services/drop_pages/`
- `core/repositories/drop_pages.go`
- `core/web/pages/drop_pages/`
- `core/web/components/drop_pages/`
- `migrations/*_create_drop_pages.sql`

## SQL Notes
- Keep SQL in repos only, uppercase.
- Use `SELECT ... FOR UPDATE` for reorder writes.
- Enforce `slug` uniqueness per owner:
  - Unique index on (owner_id, slug).

## Endpoints (Draft)
Manage:
- GET `/dashboard/pages`
- GET `/dashboard/pages/new`
- POST `/dashboard/pages`
- GET `/dashboard/pages/:id`
- POST `/dashboard/pages/:id`
- POST `/dashboard/pages/:id/items`
- POST `/dashboard/pages/:id/reorder`
- POST `/dashboard/pages/:id/delete`

Public:
- GET `/@/:username/:slug`
- GET `/p/:slug` (if using global slugs)
- POST `/@/:username/:slug/password` (if password gated)

## Validation Rules
- title: required, 3-80 chars
- slug: required, 3-64, lowercase + hyphen, no spaces
- description: optional, 0-240
- password: optional, 6-128
- files: at least 1 for publish

Validation errors live in `core/services/drop_pages/errors.go`.

## Analytics
MVP:
- Increment `view_count` on page load (unique per viewer hash per day).
- Increment download counter per file.

Later:
- Store aggregated stats per day for dashboard charts.

## Ads Integration (Phase 2)
Public page:
- Banner or native placement between file list sections.
Owner dashboard:
- Small placement in manage view.
Rewarded mechanics:
- Watch ad to unlock custom slug.
- Watch ad to remove ads for 24 hours on this page.
- Watch ad to boost download speed for 1 hour.

## MVP Build Plan
1) Migrations for `drop_pages` and `drop_page_items`.
2) Repos for CRUD + list + reorder.
3) Service for create/update/reorder with validation.
4) Handlers for manage + public view.
5) Web pages for manage and public view.
6) Basic stats counters.
7) QA: slug collision, password gate, reorder integrity.

## Open Questions
- Pick user-facing name: Drop Pages vs Share Pages.
- URL shape: `@user/slug` or global `p/slug`.
- How to handle cover image: file reference vs upload new.
- Do we allow mixing archived files or only active files.
