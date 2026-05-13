import { fetchEncryptedChunk } from "./chunk_fetcher.js";

export async function downloadBlobFallback(options) {
  const record = options.record;
  const chunkMap = options.chunkMap || [];
  const decryptChunk = options.decryptChunk;
  const filename = options.filename;
  const signal = options.signal;
  const onProgress = options.onProgress;
  const parts = [];
  let written = 0;

  for (let i = 0; i < chunkMap.length; i++) {
    const chunk = chunkMap[i];
    if (signal && signal.aborted) {
      throw new DOMException("Download aborted", "AbortError");
    }

    const encryptedBuffer = await fetchEncryptedChunk(record.sourceUrl, chunk, signal);
    const plainBuffer = await decryptChunk(chunk, encryptedBuffer);
    parts.push(plainBuffer);

    written += Number(chunk.plainSize || plainBuffer.length || 0);
    if (onProgress) {
      onProgress({
        written: written,
        total: Number(record.plaintextSize || 0),
        chunkIndex: chunk.index,
      });
    }
  }

  const blob = new Blob(parts, {
    type: record.mimeType || "application/octet-stream",
  });
  const url = URL.createObjectURL(blob);

  try {
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename || record.filename || "download";
    anchor.rel = "noopener";
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
  } finally {
    window.setTimeout(function() {
      URL.revokeObjectURL(url);
    }, 30000);
  }
}
