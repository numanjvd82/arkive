# Uploads TODO

## Not Implemented Yet
- Abuse safeguards: new account upload limits (age/verification gates) with progressive unlocks.
- Abuse safeguards: per-token and per-IP limits for public share downloads, with burst control.
- Abuse safeguards: heuristic detection for upload spikes, repeated failures, and high share churn.
- Periodic cleanup for stale multipart uploads (database records + R2 uploads/objects).
- Upload lifecycle metrics/logging (start/part/complete/abort/fail counters and durations).
- Automated tests for multipart edge cases (concurrent abort/complete, resume after partial failure, R2 transient errors).
- Integrity checks after completion (verify size/hash matches expected, fail if mismatch).
- Extract image metadata via `ffprobe` and store it alongside files.
- Generate video thumbnails during processing.
- Generate image thumbnails during processing.
- Multi-file and folder upload UI (queue, progress per item, folder support).
- Dedicated paused-uploads list in the dashboard or uploads page (beyond the resume banner).
- Alert component for inline form and page messaging.
- Add `ads.txt` for search/revenue support.
- Add "Show on Google and other search engines" option for public shares; default off, and only index when enabled.
- Add ad provider name to Privacy Policy and Cookie Policy once finalized.
- Adblock detection: deprecated (removed).
- Add governing law and dispute venue to Terms of Service after company registration.
- Add company address and phone to Terms of Service after company registration.
- Add rewards disclaimer to Privacy Policy (availability, fraud checks, subject to change).
- Inactivity retention warnings: send emails at 150/180/210 days inactive (archive/deletion notices) once email service is in place.
- Archived files UX: users may be blocked from freeing space due to 2 GB/day restore cap; add a restore CTA or archive management view.
- Drop Pages (Collections): product spec + data model + ads mechanics + MVP build plan.

## Done
- Install `ffprobe` on the server for video metadata extraction.
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
