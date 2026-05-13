export function buildChunkMap(manifest, fallbackChunkSize, plaintextSize) {
  let cipherOffset = 0;
  let plainOffset = 0;
  const totalPlaintextSize = Number(plaintextSize || 0);
  const chunkSize = Number(fallbackChunkSize || 0);

  return ((manifest && manifest.chunks) || []).map(function(chunk, index) {
    const cipherSize = Number(chunk.cipher_size ?? chunk.cipherSize ?? 0);
    let plainSize = Number(chunk.plain_size ?? chunk.plainSize ?? 0);

    if (!plainSize && chunkSize > 0) {
      plainSize = chunkSize;
      if (totalPlaintextSize > 0) {
        const remaining = totalPlaintextSize - plainOffset;
        if (remaining > 0 && remaining < plainSize) {
          plainSize = remaining;
        }
      }
    }

    const mapped = {
      index: index,
      cipherStart: cipherOffset,
      cipherEnd: cipherOffset + cipherSize - 1,
      cipherSize: cipherSize,
      plainStart: plainOffset,
      plainSize: plainSize,
      hash: chunk.hash || "",
      aad: chunk.aad || "",
    };

    cipherOffset += cipherSize;
    plainOffset += plainSize;
    return mapped;
  });
}
