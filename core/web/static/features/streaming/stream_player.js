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

  const session = await openStreamSession({
    reader: reader,
    record: record,
    metadata: metadata,
    fileId: settings.fileId,
  });

  const media = document.createElement(kind);
  media.className = settings.className || "";
  media.controls = true;
  media.preload = settings.preload || "none";
  media.setAttribute("data-video-element", "true");
  media.playsInline = true;
  media.src = createStreamUrl(record.fileId || settings.fileId || "file", session.sessionId);

  if (typeof settings.onLoadedMetadata === "function") {
    media.addEventListener("loadedmetadata", function() {
      settings.onLoadedMetadata(media);
    });
  }

  if (typeof settings.onError === "function") {
    media.addEventListener("error", function() {
      settings.onError(new Error("Encrypted stream playback failed."));
    }, { once: true });
  }

  if (typeof settings.onElement === "function") {
    settings.onElement(media);
  }

  return {
    sessionId: session.sessionId,
    element: media,
    dispose: function() {
      closeStreamSession(session.sessionId);
      try {
        media.pause();
      } catch (_) {}
      media.removeAttribute("src");
      media.load();
    },
  };
}
