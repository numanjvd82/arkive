import { fetchEncryptedChunk } from "./chunk_fetcher.js";

export async function downloadStreamedToDisk(options) {
  const record = options.record;
  const chunkMap = options.chunkMap || [];
  const decryptChunk = options.decryptChunk;
  const filename = options.filename;
  const signal = options.signal;
  const onProgress = options.onProgress;
  const readAhead = options.readAhead === true;

  if (!("showSaveFilePicker" in window) || !window.isSecureContext) {
    throw new Error("Streamed save is not supported in this browser");
  }

  const handle = await window.showSaveFilePicker({
    suggestedName: filename || record.filename || "download",
  });
  const writable = await handle.createWritable();
  let written = 0;

  try {
    if (!readAhead) {
      for (let i = 0; i < chunkMap.length; i++) {
        const chunk = chunkMap[i];
        if (signal && signal.aborted) {
          throw new DOMException("Download aborted", "AbortError");
        }

        const encryptedBuffer = await fetchEncryptedChunk(record.sourceUrl, chunk, signal);
        const plainBuffer = await decryptChunk(chunk, encryptedBuffer);
        await writable.write(plainBuffer);

        written += Number(chunk.plainSize || plainBuffer.length || 0);
        if (onProgress) {
          onProgress({
            written: written,
            total: Number(record.plaintextSize || 0),
            chunkIndex: chunk.index,
          });
        }
      }
    } else {
      let pending = prepareChunk(chunkMap, 0, record.sourceUrl, signal, decryptChunk);
      for (let i = 0; i < chunkMap.length; i++) {
        const current = await pending;
        pending = prepareChunk(chunkMap, i + 1, record.sourceUrl, signal, decryptChunk);

        const chunk = current.chunk;
        const plainBuffer = current.plainBuffer;
        await writable.write(plainBuffer);

        written += Number(chunk.plainSize || plainBuffer.length || 0);
        if (onProgress) {
          onProgress({
            written: written,
            total: Number(record.plaintextSize || 0),
            chunkIndex: chunk.index,
          });
        }
      }
    }

    await writable.close();
  } catch (error) {
    try {
      await writable.abort();
    } catch (_) {}
    throw error;
  }
}

async function prepareChunk(chunkMap, index, sourceUrl, signal, decryptChunk) {
  if (index >= chunkMap.length) {
    return null;
  }

  if (signal && signal.aborted) {
    throw new DOMException("Download aborted", "AbortError");
  }

  const chunk = chunkMap[index];
  const encryptedBuffer = await fetchEncryptedChunk(sourceUrl, chunk, signal);
  const plainBuffer = await decryptChunk(chunk, encryptedBuffer);
  return {
    chunk: chunk,
    plainBuffer: plainBuffer,
  };
}
