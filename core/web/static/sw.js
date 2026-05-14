const ARKIVE_SW_VERSION = "2026-05-14-download-v2";
const sessions = new Map();
const pendingReads = new Map();
const INITIAL_PROBE_BYTES = 4 * 1024 * 1024;
const MAX_STREAM_RESPONSE_BYTES = 4 * 1024 * 1024;
const STREAM_READ_TIMEOUT_MS = 60000;
const PREVIEW_SESSION_MAX_LIFETIME_MS = 6 * 60 * 60 * 1000;
const DOWNLOAD_SESSION_MAX_LIFETIME_MS = 45 * 60 * 1000;

self.addEventListener("install", function(event) {
  event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", function(event) {
  event.waitUntil(self.clients.claim());
});

self.addEventListener("message", function(event) {
  cleanupExpiredSessions();
  const message = event.data || {};

  if (message.type === "ARKIVE_STREAM_OPEN" && message.session) {
    const now = Date.now();
    const session = message.session;
    sessions.set(String(session.sessionId || ""), {
      sessionId: String(session.sessionId || ""),
      purpose: String(session.purpose || "preview"),
      fileId: String(session.fileId || ""),
      filename: String(session.filename || "download"),
      plaintextSize: Number(session.plaintextSize || 0),
      mimeType: String(session.mimeType || "application/octet-stream"),
      clientId: event.source && event.source.id ? String(event.source.id) : "",
      downloadOffset: 0,
      openedAt: now,
      maxLifetimeMs: sessionLifetimeMs(session),
      expiresAt: now + sessionLifetimeMs(session),
      version: ARKIVE_SW_VERSION,
    });
    return;
  }

  if (message.type === "ARKIVE_STREAM_CLOSE") {
    sessions.delete(String(message.sessionId || ""));
    return;
  }

  if (message.type === "ARKIVE_STREAM_READ_RESULT") {
    const key = String(message.sessionId || "") + ":" + String(message.requestId || "");
    const pending = pendingReads.get(key);
    if (!pending) {
      return;
    }
    pendingReads.delete(key);
    pending.resolve(new Uint8Array(message.bytes || new ArrayBuffer(0)));
    return;
  }

  if (message.type === "ARKIVE_STREAM_READ_ERROR") {
    const key = String(message.sessionId || "") + ":" + String(message.requestId || "");
    const pending = pendingReads.get(key);
    if (!pending) {
      return;
    }
    pendingReads.delete(key);
    pending.reject(new Error(String(message.error || "Stream read failed.")));
  }
});

self.addEventListener("fetch", function(event) {
  cleanupExpiredSessions();
  const url = new URL(event.request.url);
  if (!url.pathname.startsWith("/arkive-stream/") && !url.pathname.startsWith("/arkive-download/")) {
    return;
  }
  event.respondWith(
    handleRequest(event.request, url).catch(function(error) {
      return new Response(String((error && error.message) || "Encrypted stream failed."), {
        status: 504,
        headers: {
          "Cache-Control": "no-store",
          "Content-Type": "text/plain; charset=utf-8",
        },
      });
    }),
  );
});

function isDownloadRequest(url) {
  return url.pathname.startsWith("/arkive-download/");
}

function parseRange(header, size) {
  if (!header || !header.startsWith("bytes=")) {
    return {
      start: 0,
      end: Math.min(Math.max(0, size - 1), INITIAL_PROBE_BYTES - 1),
      explicit: false,
    };
  }

  const parts = header.replace("bytes=", "").split("-");
  const startRaw = parts[0];
  const endRaw = parts[1];

  if (startRaw === "" && endRaw) {
    const suffixSize = Math.max(0, Number(endRaw || 0));
    return {
      start: Math.max(0, size - suffixSize),
      end: Math.max(0, size - 1),
      explicit: true,
    };
  }

  const start = Number(startRaw || 0);
  const end = endRaw
    ? Number(endRaw)
    : Math.min(size - 1, start + MAX_STREAM_RESPONSE_BYTES - 1);

  return {
    start: Math.max(0, start),
    end: Math.min(size - 1, end),
    explicit: true,
  };
}

function clampRange(range, size) {
  const start = Math.max(0, Number(range && range.start) || 0);
  const requestedEnd = Math.min(size - 1, Math.max(start, Number(range && range.end) || start));
  const end = Math.min(requestedEnd, start + MAX_STREAM_RESPONSE_BYTES - 1);

  return {
    start: start,
    end: end,
    explicit: Boolean(range && range.explicit),
  };
}

async function handleRequest(request, url) {
  if (isDownloadRequest(url)) {
    return handleDownloadRequest(url);
  }
  return handleStreamRequest(request, url);
}

async function handleStreamRequest(request, url) {
  const sessionId = String(url.searchParams.get("session") || "");
  const session = sessions.get(sessionId);
  if (!session) {
    return new Response("Stream session not found.", { status: 404 });
  }
  if (isSessionExpired(session)) {
    finalizeSession(session, "expired");
    return new Response("Stream session expired.", { status: 410 });
  }
  touchSession(session);

  const range = clampRange(
    parseRange(request.headers.get("Range"), session.plaintextSize),
    session.plaintextSize,
  );
  if (range.end < range.start || session.plaintextSize <= 0) {
    return new Response(null, {
      status: 416,
      headers: {
        "Content-Range": "bytes */" + String(session.plaintextSize || 0),
      },
    });
  }

  const bytes = await requestBytesFromClient(session, range.start, range.end);

  return new Response(bytes, {
    status: 206,
    headers: {
      "Accept-Ranges": "bytes",
      "Cache-Control": "no-store",
      "Content-Length": String(bytes.byteLength || bytes.length || 0),
      "Content-Range": "bytes " + String(range.start) + "-" + String(range.end) + "/" + String(session.plaintextSize),
      "Content-Type": session.mimeType || "application/octet-stream",
    },
  });
}

async function handleDownloadRequest(url) {
  const sessionId = String(url.searchParams.get("session") || "");
  const session = sessions.get(sessionId);
  if (!session) {
    return new Response("Download session not found.", { status: 404 });
  }
  if (isSessionExpired(session)) {
    finalizeSession(session, "expired");
    return new Response("Download session expired.", { status: 410 });
  }
  touchSession(session);

  const headers = {
    "Cache-Control": "no-store",
    "Content-Type": session.mimeType || "application/octet-stream",
    "Content-Disposition": buildContentDisposition(session.filename || "download"),
    "Content-Length": String(session.plaintextSize || 0),
  };

  const stream = new ReadableStream({
    async pull(controller) {
      const offset = Number(session.downloadOffset || 0);
      const total = Number(session.plaintextSize || 0);
      if (offset >= total) {
        controller.close();
        finalizeSession(session, "complete");
        return;
      }

      const end = Math.min(total - 1, offset + MAX_STREAM_RESPONSE_BYTES - 1);

      try {
        const bytes = await requestBytesFromClient(session, offset, end);
        touchSession(session);
        session.downloadOffset = end + 1;
        controller.enqueue(bytes);
      } catch (error) {
        controller.error(error);
        finalizeSession(session, "error");
      }
    },
    cancel() {
      finalizeSession(session, "cancelled");
    },
  });

  return new Response(stream, {
    status: 200,
    headers: headers,
  });
}

async function requestBytesFromClient(session, start, end) {
  if (isSessionExpired(session)) {
    throw new Error("Stream session expired.");
  }
  const client = await self.clients.get(session.clientId);
  if (!client) {
    throw new Error("Stream page is unavailable.");
  }

  const requestId = String(Date.now()) + "-" + Math.random().toString(36).slice(2);
  const key = session.sessionId + ":" + requestId;

  return new Promise(function(resolve, reject) {
    pendingReads.set(key, { resolve: resolve, reject: reject });

    client.postMessage({
      type: "ARKIVE_STREAM_READ",
      sessionId: session.sessionId,
      requestId: requestId,
      start: start,
      end: end,
    });

    setTimeout(function() {
      const pending = pendingReads.get(key);
      if (!pending) {
        return;
      }
      pendingReads.delete(key);
      reject(new Error("Timed out waiting for encrypted stream data."));
    }, STREAM_READ_TIMEOUT_MS);
  });
}

function buildContentDisposition(filename) {
  const fallback = String(filename || "download").replace(/[^\x20-\x7E]+/g, "_").replace(/["\\]/g, "_");
  const encoded = encodeURIComponent(String(filename || "download")).replace(/['()]/g, escape).replace(/\*/g, "%2A");
  return 'attachment; filename="' + fallback + '"; filename*=UTF-8\'\'' + encoded;
}

function finalizeSession(session, status) {
  if (!session || !session.sessionId) {
    return;
  }
  sessions.delete(String(session.sessionId));
  notifyClient(session.clientId, {
    type: "ARKIVE_STREAM_FINISH",
    sessionId: String(session.sessionId),
    status: String(status || "complete"),
  });
}

function touchSession(session) {
  if (!session) {
    return;
  }
  session.expiresAt = Date.now() + Number(session.maxLifetimeMs || PREVIEW_SESSION_MAX_LIFETIME_MS);
}

function isSessionExpired(session) {
  return !session || Number(session.expiresAt || 0) <= Date.now();
}

function cleanupExpiredSessions() {
  sessions.forEach(function(session) {
    if (isSessionExpired(session)) {
      finalizeSession(session, "expired");
    }
  });
}

function sessionLifetimeMs(session) {
  if (session && String(session.purpose || "") === "download") {
    return DOWNLOAD_SESSION_MAX_LIFETIME_MS;
  }
  return PREVIEW_SESSION_MAX_LIFETIME_MS;
}

async function notifyClient(clientId, message) {
  if (!clientId) {
    return;
  }
  try {
    const client = await self.clients.get(clientId);
    if (client) {
      client.postMessage(message);
    }
  } catch (_) {}
}
