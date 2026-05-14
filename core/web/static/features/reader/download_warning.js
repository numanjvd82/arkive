import { getDownloadCapabilities } from "./capabilities.js";
import { canUseBlobFallback, formatBytes } from "./download_limits.js";
import { supportsServiceWorkerDownload } from "../streaming/stream_capabilities.js";

export function canDownloadInCurrentBrowser(record) {
  const caps = getDownloadCapabilities();
  return caps.supportsStreamedSave || canUseBlobFallback(record, caps) || supportsServiceWorkerDownload();
}

export function showLargeDownloadWarning(container, record) {
  if (!container) {
    return;
  }

  container.innerHTML =
    '<div class="arkive-download-warning">' +
      "<h3>Large download needs a supported browser</h3>" +
      "<p>This file is " + formatBytes(record && record.plaintextSize) + ". Your browser cannot save decrypted chunks directly to disk.</p>" +
      "<p>For large encrypted downloads, use Chrome or Edge on desktop.</p>" +
      '<p class="muted">Large encrypted downloads are limited on iPhone and iPad browsers. Streaming playback works, but saving very large decrypted files currently requires desktop Chrome/Edge or a future native Arkive app.</p>' +
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

export function showServiceWorkerDownloadNotice(container) {
  if (!container) {
    return;
  }

  container.innerHTML =
    '<div class="arkive-download-warning">' +
      "<h3>Browser download started</h3>" +
      "<p>Arkive is streaming decrypted bytes to the browser download manager.</p>" +
      '<p class="muted">Keep this tab open until the download has clearly started or completed, especially on mobile browsers.</p>' +
    "</div>";
}

export function isDownloadAbortError(error) {
  if (!error) {
    return false;
  }

  if (error.name === "AbortError") {
    return true;
  }

  const message = String(error.message || "").toLowerCase();
  return message.includes("user aborted a request");
}

export function maybeShowDownloadCapabilityWarning(root, record) {
  const container = (root || document).querySelector("#download-warning");

  if (!canDownloadInCurrentBrowser(record)) {
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
