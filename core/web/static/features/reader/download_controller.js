import { getDownloadCapabilities } from "./capabilities.js";
import { canUseBlobFallback } from "./download_limits.js";
import { clearDownloadWarning, showLargeDownloadWarning } from "./download_warning.js";
import { downloadBlobFallback } from "./download_blob.js";
import { downloadStreamedToDisk } from "./download_stream.js";

export async function downloadFile(options) {
  const record = options.record;
  const warningContainer = options.warningContainer || null;
  const caps = getDownloadCapabilities();
  const controller = new AbortController();
  const signal = options.signal || controller.signal;

  clearDownloadWarning(warningContainer);

  if (caps.supportsStreamedSave) {
    await downloadStreamedToDisk({
      record: record,
      chunkMap: options.chunkMap,
      decryptChunk: options.decryptChunk,
      filename: options.filename,
      signal: signal,
      onProgress: options.onProgress,
    });
    return { mode: "stream" };
  }

  if (canUseBlobFallback(record, caps)) {
    await downloadBlobFallback({
      record: record,
      chunkMap: options.chunkMap,
      decryptChunk: options.decryptChunk,
      filename: options.filename,
      signal: signal,
      onProgress: options.onProgress,
    });
    return { mode: "blob" };
  }

  if (warningContainer) {
    showLargeDownloadWarning(warningContainer, record);
    return { mode: "warning" };
  }

  throw new Error("This browser cannot save decrypted chunks directly to disk. Use Chrome or Edge on desktop for large encrypted downloads.");
}
