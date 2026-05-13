const sessions = new Map();
const pendingReads = new Map();
const INITIAL_PROBE_BYTES = 4 * 1024 * 1024;
const MAX_STREAM_RESPONSE_BYTES = 4 * 1024 * 1024;
const STREAM_READ_TIMEOUT_MS = 60000;

self.addEventListener("install", function(event) {
  event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", function(event) {
  event.waitUntil(self.clients.claim());
});

self.addEventListener("message", function(event) {
  const message = event.data || {};

  if (message.type === "ARKIVE_STREAM_OPEN" && message.session) {
    const session = message.session;
    sessions.set(String(session.sessionId || ""), {
      sessionId: String(session.sessionId || ""),
      fileId: String(session.fileId || ""),
      plaintextSize: Number(session.plaintextSize || 0),
      mimeType: String(session.mimeType || "application/octet-stream"),
      clientId: event.source && event.source.id ? String(event.source.id) : "",
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
  const url = new URL(event.request.url);
  if (!url.pathname.startsWith("/arkive-stream/")) {
    return;
  }
  event.respondWith(
    handleStreamRequest(event.request, url).catch(function(error) {
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

async function handleStreamRequest(request, url) {
  const sessionId = String(url.searchParams.get("session") || "");
  const session = sessions.get(sessionId);
  if (!session) {
    return new Response("Stream session not found.", { status: 404 });
  }

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

async function requestBytesFromClient(session, start, end) {
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
