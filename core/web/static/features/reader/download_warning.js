import { getDownloadCapabilities } from "./capabilities.js";
import { canUseBlobFallback, formatBytes } from "./download_limits.js";

export function showLargeDownloadWarning(container, record) {
  if (!container) {
    return;
  }

  container.innerHTML =
    '<div class="arkive-download-warning">' +
      "<h3>Large download needs a supported browser</h3>" +
      "<p>This file is " + formatBytes(record && record.plaintextSize) + ". Your browser cannot save decrypted chunks directly to disk.</p>" +
      "<p>For large encrypted downloads, use Chrome or Edge on desktop.</p>" +
      '<p class="muted">Safari, iOS Safari, and Firefox may only support smaller downloads because they require Arkive to decrypt the file into browser memory first.</p>' +
    "</div>";
}

export function clearDownloadWarning(container) {
  if (!container) {
    return;
  }
  container.innerHTML = "";
}

export function showDownloadError(container, message) {
  if (!container) {
    return;
  }

  container.innerHTML =
    '<div class="arkive-download-warning">' +
      "<h3>Download failed</h3>" +
      "<p>" + escapeHTML(message || "Something went wrong.") + "</p>" +
    "</div>";
}

export function maybeShowDownloadCapabilityWarning(root, record) {
  const container = (root || document).querySelector("#download-warning");
  const caps = getDownloadCapabilities();

  if (!caps.supportsStreamedSave && !canUseBlobFallback(record, caps)) {
    showLargeDownloadWarning(container, record);
    return;
  }

  clearDownloadWarning(container);
}

function escapeHTML(value) {
  return String(value).replace(/[&<>"']/g, function(ch) {
    return {
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#039;",
    }[ch];
  });
}
