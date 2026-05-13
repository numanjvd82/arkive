export function getDownloadCapabilities() {
  const supportsStreamedSave =
    typeof window !== "undefined" &&
    "showSaveFilePicker" in window &&
    window.isSecureContext;

  const nav = typeof navigator !== "undefined" ? navigator : null;
  const ua = nav && nav.userAgent ? nav.userAgent : "";
  const platform = nav && nav.platform ? nav.platform : "";
  const maxTouchPoints = nav && nav.maxTouchPoints ? nav.maxTouchPoints : 0;

  const isIOS =
    /iPad|iPhone|iPod/.test(ua) ||
    (platform === "MacIntel" && maxTouchPoints > 1);

  const isSafari =
    /^((?!chrome|android|crios|fxios|edg|opr).)*safari/i.test(ua);

  const isFirefox = /firefox|fxios/i.test(ua);

  return {
    supportsStreamedSave,
    isIOS,
    isSafari,
    isFirefox,
    shouldWarnForLargeDownloads: !supportsStreamedSave,
  };
}
