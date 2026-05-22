import { toAppError } from "./errors.js";

function titleForCode(code) {
  switch (String(code || "")) {
    case "storage_limit_exceeded":
      return "Storage limit reached";
    case "upload_failed":
      return "Upload failed";
    case "download_failed":
      return "Download failed";
    case "thumbnail_failed":
      return "Thumbnail failed";
    case "validation_failed":
      return "Validation failed";
    default:
      return "Error";
  }
}

export function showAppError(error, fallback) {
  const appError = toAppError(error, fallback);
  if (appError.name === "AbortError") {
    return appError;
  }
  if (window.Toast) {
    window.Toast.error(appError.message, { title: titleForCode(appError.code) });
  }
  return appError;
}
