# Uploads TODO

## Not Implemented Yet
- Periodic cleanup for stale multipart uploads (database records + R2 uploads/objects).
- Upload lifecycle metrics/logging (start/part/complete/abort/fail counters and durations).
- Generate video thumbnails during processing.
- Generate image thumbnails during processing.
- Dedicated paused-uploads list in the dashboard or uploads page (beyond the resume banner).
- Burn-after-reading share policy in Core (single-download enforcement, backend state, public share flow, UI wiring).
- Inactivity retention warnings: send emails at 150/180/210 days inactive (archive/deletion notices) once email service is in place.
- CSP hardening follow-up: keep `connect-src`/`img-src`/`media-src` permissive by default in Core because storage origins are user-defined; add config-driven CSP origin allowlists later for hardened self-hosted deployments.
- Remove backend pagination and use db pagination instead.
- Add grid/list view
- Add folder capabilities so we can move files to it.
- The ability to create simple text files inside the file manager
- Add Exif toggle in the settings and show a dialog to strip exif or keep it before uploading.
- Add session management to the settings. 
- Process files in a queue like one by one => Priorty: HIGH

## Done
- Install `ffprobe` on the server for video metadata extraction.
- Replace placeholder setup recovery key with real browser-generated key from `arkive-crypto` WASM; key material must stay client-side and no plaintext recovery phrase should be sent to the server.
- Rate limiting for upload endpoints and background media processing.
- Add Cloudflare analytics script to main production layout.
- Refactor media view page (components, helpers) for better structure and maintainability.
- Strategy for private/public view and download expiry (stable media URLs, refresh flow, tokenized CDN).
- Improve toast visibility and messaging for rate-limited actions across the app.
- SEO improvements (titles, descriptions, canonical URLs, sitemap, JSON-LD).
- Add `robots.txt` for search/revenue support.
- Add Open Graph/Twitter meta tags for link previews.
- Add sitemap.xml for marketing pages.
- Refactor share UI (dialog + public share view) for clearer components and state management.
- Add cross-links between Privacy Policy, Cookie Policy, Terms of Service, and AUP.
- Add revision dates to Terms of Service and AUP pages.
- Add SEO landing pages: /secure-file-sharing, /share-large-files, /file-sharing-without-login.
- Add internal links from home/pricing/footer to the new SEO landing pages.
- Competitive matrix: Arkive vs Drive/Dropbox/iCloud/OneDrive/TeraBox with feature comparisons, “why Arkive” messaging, and homepage/pricing bullets.
