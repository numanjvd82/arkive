export const STATUS = Object.freeze({
	QUEUED: "queued",
	RUNNING: "running",
	COMPLETED: "completed",
	FAILED: "failed",
	CANCELED: "canceled",
});

export function isTerminal(status) {
	return status === STATUS.COMPLETED || status === STATUS.FAILED || status === STATUS.CANCELED;
}

export function isActive(status) {
	return status === STATUS.QUEUED || status === STATUS.RUNNING;
}
