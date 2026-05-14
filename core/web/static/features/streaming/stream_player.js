import { canStreamMedia, mediaKindForMime } from "./stream_capabilities.js";
import { closeStreamSession, openStreamSession } from "./stream_session.js";
import { createStreamUrl } from "./stream_url.js";

export async function mountStreamingMedia(options) {
  const settings = options || {};
  const reader = settings.reader;
  const metadata = settings.metadata || {};
  const record = settings.record || {};
  const mime = String(metadata.mime || record.mimeType || "").toLowerCase();
  const kind = String(settings.kind || mediaKindForMime(mime) || "video");

  if (!canStreamMedia(mime)) {
    throw new Error("This browser cannot stream this encrypted media format.");
  }

  const media = document.createElement(kind);
  media.className = settings.className || "";
  media.controls = true;
  media.preload = settings.preload || "none";
  media.setAttribute("data-video-element", "true");
  media.playsInline = true;
  let session = null;
  let disposed = false;
  let recovering = false;
  let recoveryAttempts = 0;
  const maxRecoveryAttempts = Number(settings.maxRecoveryAttempts || 2);

  async function openSession() {
    return openStreamSession({
      reader: reader,
      record: record,
      metadata: metadata,
      fileId: settings.fileId,
    });
  }

  async function attachSession(nextSession, resumeTime) {
    session = nextSession;
    media.src = createStreamUrl(record.fileId || settings.fileId || "file", session.sessionId);
    media.load();

    if (typeof resumeTime === "number" && resumeTime > 0) {
      const restoreTime = function() {
        media.removeEventListener("loadedmetadata", restoreTime);
        try {
          media.currentTime = resumeTime;
        } catch (_) {}
      };
      media.addEventListener("loadedmetadata", restoreTime);
    }
  }

  async function recoverPlayback() {
    if (disposed || recovering || recoveryAttempts >= maxRecoveryAttempts) {
      return false;
    }

    recovering = true;
    recoveryAttempts += 1;

    const resumeTime = Number(media.currentTime || 0);
    const wasPaused = media.paused;
    const previousSessionId = session && session.sessionId ? session.sessionId : "";

    try {
      const nextSession = await openSession();
      if (previousSessionId) {
        closeStreamSession(previousSessionId);
      }
      await attachSession(nextSession, resumeTime);
      if (!wasPaused) {
        media.play().catch(function() {});
      }
      recovering = false;
      return true;
    } catch (_) {
      recovering = false;
      return false;
    }
  }

  session = await openSession();
  await attachSession(session);

  if (typeof settings.onLoadedMetadata === "function") {
    media.addEventListener("loadedmetadata", function() {
      settings.onLoadedMetadata(media);
    });
  }

  if (typeof settings.onError === "function") {
    media.addEventListener("error", function() {
      recoverPlayback().then(function(recovered) {
        if (!recovered && !disposed) {
          settings.onError(new Error("Encrypted stream playback failed."));
        }
      });
    });
  }

  if (typeof settings.onElement === "function") {
    settings.onElement(media);
  }

  return {
    sessionId: session.sessionId,
    element: media,
    dispose: function() {
      disposed = true;
      if (session && session.sessionId) {
        closeStreamSession(session.sessionId);
      }
      try {
        media.pause();
      } catch (_) {}
      media.removeAttribute("src");
      media.load();
    },
  };
}
