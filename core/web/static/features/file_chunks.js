const MB = 1024 * 1024;
const GB = 1024 * MB;

export const FILE_CHUNK_POLICY = Object.freeze({
  smallFileMaxBytes: 64 * MB,
  largeFileMinBytes: 1 * GB,
  smallUploadChunkBytes: 1 * MB,
  defaultUploadChunkBytes: 4 * MB,
  largeUploadChunkBytes: 8 * MB,
  maxLargeUploadChunkBytes: 16 * MB,
  videoPreviewChunkMinBytes: 1 * MB,
  videoPreviewChunkMaxBytes: 4 * MB
});

export function resolveUploadChunkSize(fileSize) {
  if (fileSize < FILE_CHUNK_POLICY.smallFileMaxBytes) {
    return FILE_CHUNK_POLICY.smallUploadChunkBytes;
  }
  if (fileSize > FILE_CHUNK_POLICY.largeFileMinBytes) {
    return FILE_CHUNK_POLICY.largeUploadChunkBytes;
  }
  return FILE_CHUNK_POLICY.defaultUploadChunkBytes;
}
