import { getDownloadCapabilities } from "./capabilities.js";
import { DOWNLOAD_POLICY } from "./download_policy.js";

export const DESKTOP_BLOB_DOWNLOAD_LIMIT_BYTES = DOWNLOAD_POLICY.blobFallbackLimitDesktop;
export const MOBILE_BLOB_DOWNLOAD_LIMIT_BYTES = DOWNLOAD_POLICY.blobFallbackLimitIOS;

export function blobDownloadLimitBytes(capabilities) {
  const caps = capabilities || getDownloadCapabilities();
  if (caps.isIOS) {
    return MOBILE_BLOB_DOWNLOAD_LIMIT_BYTES;
  }
  return DESKTOP_BLOB_DOWNLOAD_LIMIT_BYTES;
}

export function canUseBlobFallback(record, capabilities) {
  return Number(record && record.plaintextSize ? record.plaintextSize : 0) <= blobDownloadLimitBytes(capabilities);
}

export function formatBytes(bytes) {
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = Number(bytes || 0);
  let index = 0;

  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }

  return String(value.toFixed(value >= 10 || index === 0 ? 0 : 1)) + " " + units[index];
}
