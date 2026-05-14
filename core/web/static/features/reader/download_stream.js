import { fetchEncryptedChunk } from "./chunk_fetcher.js";

export async function downloadStreamedToDisk(options) {
  const record = options.record;
  const chunkMap = options.chunkMap || [];
  const decryptChunk = options.decryptChunk;
  const filename = options.filename;
  const signal = options.signal;
  const onProgress = options.onProgress;
  const readAhead = options.readAhead === true;
  const prepareConcurrency = Math.min(
    3,
    Math.max(1, Number(options.prepareConcurrency || 1)),
  );

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
      const prepared = new Map();
      let nextIndexToSchedule = 0;

      function fillPreparationQueue() {
        while (nextIndexToSchedule < chunkMap.length && prepared.size < prepareConcurrency) {
          prepared.set(
            nextIndexToSchedule,
            prepareChunk(chunkMap, nextIndexToSchedule, record.sourceUrl, signal, decryptChunk),
          );
          nextIndexToSchedule += 1;
        }
      }

      fillPreparationQueue();

      for (let i = 0; i < chunkMap.length; i++) {
        const current = await prepared.get(i);
        prepared.delete(i);
        fillPreparationQueue();

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
