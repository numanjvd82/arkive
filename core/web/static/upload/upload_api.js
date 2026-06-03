import { apiRequest } from "../lib/api.js";

export async function getUploadLimits(signal) {
	return apiRequest("/api/uploads/limits", {
		method: "GET",
		credentials: "include",
		signal: signal,
	}, { code: "upload_limits_failed", message: "Upload limits load failed" });
}

export async function startUpload(payload, signal) {
	return apiRequest("/api/uploads/start", {
		method: "POST",
		credentials: "include",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
		signal: signal,
	}, { code: "upload_failed", message: "Upload start failed" });
}

export async function presignUploadPart(sessionId, partNumber, signal) {
	return apiRequest("/api/uploads/" + encodeURIComponent(sessionId) + "/parts/" + encodeURIComponent(String(partNumber)) + "/presign", {
		method: "POST",
		credentials: "include",
		signal: signal,
	}, { code: "upload_failed", message: "Part presign failed" });
}

export async function presignUploadParts(sessionId, partNumbers, signal) {
	return apiRequest("/api/uploads/" + encodeURIComponent(sessionId) + "/parts/presign", {
		method: "POST",
		credentials: "include",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ parts: partNumbers || [] }),
		signal: signal,
	}, { code: "upload_failed", message: "Part presign failed" });
}

export async function recordUploadPart(sessionId, payload, signal) {
	return apiRequest("/api/uploads/" + encodeURIComponent(sessionId) + "/parts", {
		method: "POST",
		credentials: "include",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
		signal: signal,
	}, { code: "upload_failed", message: "Part record failed" });
}

export async function completeUpload(sessionId, payload, signal) {
	return apiRequest("/api/uploads/" + encodeURIComponent(sessionId) + "/complete", {
		method: "POST",
		credentials: "include",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
		signal: signal,
	}, { code: "upload_failed", message: "Upload complete failed" });
}

export async function presignThumbnailUpload(sessionId, payload, signal) {
	return apiRequest("/api/uploads/" + encodeURIComponent(sessionId) + "/thumbnail/presign", {
		method: "POST",
		credentials: "include",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
		signal: signal,
	}, { code: "thumbnail_failed", message: "Thumbnail presign failed" });
}

export async function cancelUpload(sessionId) {
	await apiRequest("/api/uploads/" + encodeURIComponent(sessionId) + "/cancel", {
		method: "POST",
		credentials: "include",
	}, { code: "upload_failed", message: "Upload cancel failed" });
	return { ok: true };
}

export function cancelUploadBestEffort(sessionId) {
	if (!sessionId) return;
	try {
		fetch("/api/uploads/" + encodeURIComponent(sessionId) + "/cancel", {
			method: "POST",
			credentials: "include",
			keepalive: true,
		}).catch(function () {});
	} catch (_) {}
}
