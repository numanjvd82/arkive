const THUMBNAIL_MAX_DIMENSION = 320;
const THUMBNAIL_QUALITY = 0.7;

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

export async function generateUploadThumbnail(file) {
	if (!file || !file.type || String(file.type).toLowerCase().indexOf("image/") !== 0) {
		return null;
	}
	try {
		return await renderImageThumbnail(file);
	} catch (_) {
		return null;
	}
}
