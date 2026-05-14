async function readJSON(response, fallback) {
	const text = await response.text();
	let data = null;
	try {
		data = text ? JSON.parse(text) : null;
	} catch (_) {}
	if (!response.ok) {
		if (data && data.errors) {
			const keys = Object.keys(data.errors);
			if (keys.length > 0) {
				throw new Error(String(data.errors[keys[0]] || fallback));
			}
		}
		throw new Error((data && data.error) || text || fallback);
	}
	return data;
}

export async function startUpload(payload, signal) {
	const response = await fetch("/api/uploads/start", {
		method: "POST",
		credentials: "include",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
		signal: signal,
	});
	return readJSON(response, "Upload start failed");
}

export async function presignUploadPart(sessionId, partNumber, signal) {
	const response = await fetch("/api/uploads/" + encodeURIComponent(sessionId) + "/parts/" + encodeURIComponent(String(partNumber)) + "/presign", {
		method: "POST",
		credentials: "include",
		signal: signal,
	});
	return readJSON(response, "Part presign failed");
}

export async function presignUploadParts(sessionId, partNumbers, signal) {
	const response = await fetch("/api/uploads/" + encodeURIComponent(sessionId) + "/parts/presign", {
		method: "POST",
		credentials: "include",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ parts: partNumbers || [] }),
		signal: signal,
	});
	return readJSON(response, "Part presign failed");
}

export async function recordUploadPart(sessionId, payload, signal) {
	const response = await fetch("/api/uploads/" + encodeURIComponent(sessionId) + "/parts", {
		method: "POST",
		credentials: "include",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
		signal: signal,
	});
	return readJSON(response, "Part record failed");
}

export async function completeUpload(sessionId, payload, signal) {
	const response = await fetch("/api/uploads/" + encodeURIComponent(sessionId) + "/complete", {
		method: "POST",
		credentials: "include",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
		signal: signal,
	});
	return readJSON(response, "Upload complete failed");
}

export async function cancelUpload(sessionId) {
	const response = await fetch("/api/uploads/" + encodeURIComponent(sessionId) + "/cancel", {
		method: "POST",
		credentials: "include",
	});
	if (!response.ok) {
		throw new Error("Upload cancel failed");
	}
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
