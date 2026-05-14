import { createDownloadUrl, createStreamSessionId } from "./stream_url.js";

const providers = new Map();
let providerBridgeBound = false;

function controllerMessageTarget() {
  if (navigator.serviceWorker.controller) {
    return navigator.serviceWorker.controller;
  }
  return null;
}

function postToServiceWorker(message, transfer) {
  const target = controllerMessageTarget();
  if (!target) {
    throw new Error("Streaming service worker is not controlling this page yet.");
  }
  target.postMessage(message, transfer || []);
}

function bindProviderBridge() {
  if (providerBridgeBound || !("serviceWorker" in navigator)) {
    return;
  }
  providerBridgeBound = true;

  navigator.serviceWorker.addEventListener("message", function(event) {
    const message = event.data || {};
    if (message.type === "ARKIVE_STREAM_FINISH") {
      providers.delete(String(message.sessionId || ""));
      return;
    }
    if (message.type !== "ARKIVE_STREAM_READ") {
      return;
    }

    const provider = providers.get(String(message.sessionId || ""));
    if (!provider) {
      try {
        postToServiceWorker({
          type: "ARKIVE_STREAM_READ_ERROR",
          sessionId: String(message.sessionId || ""),
          requestId: String(message.requestId || ""),
          error: "Stream session is unavailable.",
        });
      } catch (_) {}
      return;
    }

    provider.reader.readRange(Number(message.start || 0), Number(message.end || 0) + 1)
      .then(function(bytes) {
        postToServiceWorker({
          type: "ARKIVE_STREAM_READ_RESULT",
          sessionId: provider.sessionId,
          requestId: String(message.requestId || ""),
          bytes: bytes.buffer,
        }, [bytes.buffer]);
      })
      .catch(function(error) {
        postToServiceWorker({
          type: "ARKIVE_STREAM_READ_ERROR",
          sessionId: provider.sessionId,
          requestId: String(message.requestId || ""),
          error: (error && error.message) || "Stream read failed.",
        });
      });
  });
}

export function waitForServiceWorkerController(timeoutMs) {
  if (!("serviceWorker" in navigator)) {
    return Promise.reject(new Error("Service worker is not supported in this browser."));
  }
  if (navigator.serviceWorker.controller) {
    return Promise.resolve(navigator.serviceWorker.controller);
  }

  const timeout = Number(timeoutMs || 8000);

  return new Promise(function(resolve, reject) {
    let settled = false;
    let timer = 0;

    function finish(error) {
      if (settled) {
        return;
      }
      settled = true;
      navigator.serviceWorker.removeEventListener("controllerchange", onControllerChange);
      if (timer) {
        window.clearTimeout(timer);
      }
      if (error) {
        reject(error);
        return;
      }
      resolve(navigator.serviceWorker.controller);
    }

    function onControllerChange() {
      if (navigator.serviceWorker.controller) {
        finish(null);
      }
    }

    navigator.serviceWorker.addEventListener("controllerchange", onControllerChange);
    timer = window.setTimeout(function() {
      finish(new Error("Streaming service worker is not ready yet. Reload the page and try again."));
    }, timeout);

    navigator.serviceWorker.getRegistration("/")
      .then(function(registration) {
        if (registration && registration.active) {
          return;
        }
        return navigator.serviceWorker.register("/sw.js");
      })
      .catch(function(error) {
        finish(error instanceof Error ? error : new Error("Service worker registration failed."));
      });
  });
}

export async function openStreamSession(options) {
  const settings = options || {};
  const reader = settings.reader;
  const record = settings.record || {};
  const metadata = settings.metadata || {};
  if (!reader) {
    throw new Error("Missing reader for stream session.");
  }

  bindProviderBridge();
  await waitForServiceWorkerController();

  const sessionId = createStreamSessionId();
  const session = {
    sessionId: sessionId,
    purpose: String(settings.purpose || "preview"),
    fileId: String(record.fileId || settings.fileId || "file"),
    filename: String(settings.filename || metadata.name || record.filename || "download"),
    plaintextSize: Number(metadata.size || record.plaintextSize || 0),
    mimeType: String(metadata.mime || record.mimeType || "application/octet-stream"),
  };

  providers.set(sessionId, {
    sessionId: sessionId,
    reader: reader,
  });

  try {
    postToServiceWorker({
      type: "ARKIVE_STREAM_OPEN",
      session: session,
    });
  } catch (error) {
    providers.delete(sessionId);
    throw error;
  }

  return session;
}

export function closeStreamSession(sessionId) {
  const key = String(sessionId || "");
  providers.delete(key);
  try {
    postToServiceWorker({
      type: "ARKIVE_STREAM_CLOSE",
      sessionId: key,
    });
  } catch (_) {}
}

export async function startServiceWorkerDownload(options) {
  const settings = options || {};
  const session = await openStreamSession(Object.assign({}, settings, {
    purpose: "download",
  }));
  const anchor = document.createElement("a");
  anchor.href = createDownloadUrl(session.fileId, session.sessionId);
  anchor.download = session.filename || "download";
  anchor.rel = "noopener";
  anchor.style.display = "none";
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  return session;
}
