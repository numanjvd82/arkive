const THUMBNAIL_MAX_DIMENSION = 320;
const THUMBNAIL_QUALITY = 0.7;
const VIDEO_THUMBNAIL_TIME_SECONDS = 1;
const VIDEO_THUMBNAIL_RATIO = 0.1;
const VIDEO_THUMBNAIL_LONG_DURATION_SECONDS = 30 * 60;
const VIDEO_THUMBNAIL_LONG_RATIO = 0.5;

function canvasToBlob(canvas, type, quality) {
	return new Promise(function(resolve, reject) {
		canvas.toBlob(function(blob) {
			if (!blob) {
				reject(new Error("Thumbnail encoding failed"));
				return;
			}
			resolve(blob);
		}, type, quality);
	});
}

function fitWithin(width, height, maxDimension) {
	if (width <= 0 || height <= 0) {
		return { width: 0, height: 0 };
	}
	if (width <= maxDimension && height <= maxDimension) {
		return { width: Math.max(1, Math.round(width)), height: Math.max(1, Math.round(height)) };
	}
	const scale = Math.min(maxDimension / width, maxDimension / height);
	return {
		width: Math.max(1, Math.round(width * scale)),
		height: Math.max(1, Math.round(height * scale)),
	};
}

async function renderImageThumbnail(file) {
	let bitmap = null;
	let objectURL = "";
	try {
		if (typeof createImageBitmap === "function") {
			bitmap = await createImageBitmap(file);
		} else {
			objectURL = URL.createObjectURL(file);
			bitmap = await new Promise(function(resolve, reject) {
				const image = new Image();
				image.onload = function() {
					resolve(image);
				};
				image.onerror = function() {
					reject(new Error("Image load failed"));
				};
				image.src = objectURL;
			});
		}

		const target = fitWithin(bitmap.width || 0, bitmap.height || 0, THUMBNAIL_MAX_DIMENSION);
		if (target.width <= 0 || target.height <= 0) {
			return null;
		}
		const canvas = document.createElement("canvas");
		canvas.width = target.width;
		canvas.height = target.height;
		const context = canvas.getContext("2d", { alpha: false });
		if (!context) {
			return null;
		}
		context.drawImage(bitmap, 0, 0, target.width, target.height);
		const blob = await canvasToBlob(canvas, "image/webp", THUMBNAIL_QUALITY);
		return {
			bytes: new Uint8Array(await blob.arrayBuffer()),
			mime: "image/webp",
			width: target.width,
			height: target.height,
		};
	} finally {
		if (bitmap && typeof bitmap.close === "function") {
			bitmap.close();
		}
		if (objectURL) {
			URL.revokeObjectURL(objectURL);
		}
	}
}

function waitForEvent(target, eventName, errorName) {
	return new Promise(function(resolve, reject) {
		function cleanup() {
			target.removeEventListener(eventName, onSuccess);
			if (errorName) {
				target.removeEventListener(errorName, onError);
			}
		}
		function onSuccess() {
			cleanup();
			resolve();
		}
		function onError() {
			cleanup();
			reject(new Error("Thumbnail media load failed"));
		}
		target.addEventListener(eventName, onSuccess, { once: true });
		if (errorName) {
			target.addEventListener(errorName, onError, { once: true });
		}
	});
}

async function seekVideo(video, timeSeconds) {
	if (!Number.isFinite(timeSeconds) || timeSeconds < 0) {
		return;
	}
	const clamped = Math.max(0, Math.min(timeSeconds, Math.max(0, Number(video.duration || 0))));
	if (Math.abs(Number(video.currentTime || 0) - clamped) < 0.05) {
		return;
	}
	const seeked = waitForEvent(video, "seeked", "error");
	video.currentTime = clamped;
	await seeked;
}

async function renderVideoThumbnail(file) {
	const objectURL = URL.createObjectURL(file);
	const video = document.createElement("video");
	video.preload = "metadata";
	video.muted = true;
	video.playsInline = true;
	video.crossOrigin = "anonymous";
	video.src = objectURL;
	try {
		await waitForEvent(video, "loadedmetadata", "error");
		if (!video.videoWidth || !video.videoHeight) {
			await waitForEvent(video, "loadeddata", "error");
		}
		const duration = Math.max(0, Number(video.duration || 0));
		const targetTime = duration >= VIDEO_THUMBNAIL_LONG_DURATION_SECONDS
			? duration * VIDEO_THUMBNAIL_LONG_RATIO
			: Math.min(
				VIDEO_THUMBNAIL_TIME_SECONDS,
				duration * VIDEO_THUMBNAIL_RATIO,
			);
		await seekVideo(video, targetTime);

		const target = fitWithin(video.videoWidth || 0, video.videoHeight || 0, THUMBNAIL_MAX_DIMENSION);
		if (target.width <= 0 || target.height <= 0) {
			return null;
		}

		const canvas = document.createElement("canvas");
		canvas.width = target.width;
		canvas.height = target.height;
		const context = canvas.getContext("2d", { alpha: false });
		if (!context) {
			return null;
		}
		context.drawImage(video, 0, 0, target.width, target.height);
		const blob = await canvasToBlob(canvas, "image/webp", THUMBNAIL_QUALITY);
		return {
			bytes: new Uint8Array(await blob.arrayBuffer()),
			mime: "image/webp",
			width: target.width,
			height: target.height,
		};
	} finally {
		video.pause();
		video.removeAttribute("src");
		video.load();
		URL.revokeObjectURL(objectURL);
	}
}

export async function generateUploadThumbnail(file) {
	if (!file || !file.type) {
		return null;
	}
	const mime = String(file.type).toLowerCase();
	try {
		if (mime.indexOf("image/") === 0) {
			return await renderImageThumbnail(file);
		}
		if (mime === "video/mp4" || mime === "video/webm") {
			return await renderVideoThumbnail(file);
		}
		return null;
	} catch (_) {
		return null;
	}
}
