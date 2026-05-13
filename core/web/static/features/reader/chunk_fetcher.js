export async function fetchEncryptedChunk(sourceUrl, chunk, signal) {
  const res = await fetch(sourceUrl, {
    headers: {
      Range: "bytes=" + String(chunk.cipherStart) + "-" + String(chunk.cipherEnd),
    },
    signal: signal,
  });

  if (res.status !== 206) {
    throw new Error("Expected 206 Partial Content, got " + String(res.status));
  }

  return new Uint8Array(await res.arrayBuffer());
}
