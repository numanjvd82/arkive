export const UPLOAD_POLICY = Object.freeze({
	maxActiveFiles: 1,
	fileChunkSize: 4 * 1024 * 1024,
	uploadPartSize: 8 * 1024 * 1024,
	partConcurrency: 3,
	presignBatchSize: 8,
	encryptReadAhead: 1,
});
