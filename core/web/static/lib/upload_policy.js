export const UPLOAD_POLICY = Object.freeze({
	maxActiveFiles: 1,
	largeFileThreshold: 256 * 1024 * 1024,
	defaultPartSize: 4 * 1024 * 1024,
	largeFilePartConcurrency: 2,
	smallFilePartConcurrency: 3,
});

export function getPartConcurrency(fileSize) {
	return fileSize >= UPLOAD_POLICY.largeFileThreshold
		? UPLOAD_POLICY.largeFilePartConcurrency
		: UPLOAD_POLICY.smallFilePartConcurrency;
}
