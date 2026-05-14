function getMobilePlatformContext() {
  const nav = typeof navigator !== "undefined" ? navigator : null;
  const ua = nav && nav.userAgent ? nav.userAgent : "";
  const platform = nav && nav.platform ? nav.platform : "";
  const maxTouchPoints = nav && nav.maxTouchPoints ? nav.maxTouchPoints : 0;

  const isIOS =
    /iPad|iPhone|iPod/.test(ua) ||
    (platform === "MacIntel" && maxTouchPoints > 1);

  return {
    isIOS: isIOS,
  };
}

export function supportsServiceWorkerStreaming() {
  return typeof window !== "undefined" &&
    window.isSecureContext &&
    "serviceWorker" in navigator &&
    typeof navigator.serviceWorker.addEventListener === "function";
}

export function supportsServiceWorkerDownload() {
  const context = getMobilePlatformContext();
  return supportsServiceWorkerStreaming() && !context.isIOS;
}

export function mediaKindForMime(mime) {
  const value = String(mime || "").toLowerCase();
  if (value.startsWith("video/")) {
    return "video";
  }
  if (value.startsWith("audio/")) {
    return "audio";
  }
  return "";
}

export function browserCanPlayMime(mime, kind) {
  const mediaKind = kind || mediaKindForMime(mime);
  if (!mediaKind || typeof document === "undefined") {
    return false;
  }
  const element = document.createElement(mediaKind);
  if (!element || typeof element.canPlayType !== "function") {
    return false;
  }
  return element.canPlayType(String(mime || "")).replace(/^no$/i, "") !== "";
}

export function canStreamMedia(mime) {
  const kind = mediaKindForMime(mime);
  return supportsServiceWorkerStreaming() && browserCanPlayMime(mime, kind);
}
