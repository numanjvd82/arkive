export function createStreamSessionId() {
  if (window.crypto && window.crypto.randomUUID) {
    return window.crypto.randomUUID();
  }
  return "stream-" + Date.now() + "-" + Math.random().toString(36).slice(2);
}

export function createStreamUrl(fileId, sessionId) {
  return "/arkive-stream/" +
    encodeURIComponent(String(fileId || "file")) +
    "?session=" +
    encodeURIComponent(String(sessionId || ""));
}
