import initArkiveCrypto, * as arkiveCrypto from "../vendor/arkive-crypto/arkive_crypto.js";

let cryptoReady = null;
let vaultSession = null;
let unlockedMasterKey = null;
const activeUploads = new Map();

function encodeBase64(bytes) {
	let binary = "";
	for (let i = 0; i < bytes.length; i++) {
		binary += String.fromCharCode(bytes[i]);
	}
	return btoa(binary);
}

function decodeBase64(value) {
	const normalized = String(value || "").trim();
	if (!normalized) {
		return new Uint8Array();
	}
	const binary = atob(normalized);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}
	return bytes;
}

function toUint8Array(value) {
	if (value instanceof Uint8Array) return value;
	if (value instanceof ArrayBuffer) return new Uint8Array(value);
	if (ArrayBuffer.isView(value)) {
		return new Uint8Array(value.buffer, value.byteOffset, value.byteLength);
	}
	return decodeBase64(value);
}

function aadBytes(value) {
	if (!value) return new Uint8Array();
	if (value instanceof Uint8Array) return value;
	return new TextEncoder().encode(String(value));
}

function ensureCrypto() {
	if (cryptoReady) return cryptoReady;
	cryptoReady = initArkiveCrypto({ module_or_path: "/static/vendor/arkive-crypto/arkive_crypto_bg.wasm" }).then(function () {
		return arkiveCrypto;
	});
	return cryptoReady;
}

function zeroizeAndClear(crypto) {
	if (unlockedMasterKey) {
		crypto.zeroize(unlockedMasterKey);
		unlockedMasterKey = null;
	}
	activeUploads.forEach(function (entry) {
		if (entry && entry.fileKey) {
			crypto.zeroize(entry.fileKey);
		}
	});
	activeUploads.clear();
}

export function setVaultSession(session) {
	vaultSession = session || null;
}

export async function restoreVaultSession() {
	if (!vaultSession) return;
	const crypto = await ensureCrypto();
	const sessionUnwrapKey = toUint8Array(vaultSession.sessionUnwrapKey);
	const wrappedMasterKey = decodeBase64(vaultSession.wrappedMasterKey);
	try {
		const masterKey = crypto.unwrap_master_key(wrappedMasterKey, sessionUnwrapKey, aadBytes("arkive:session-master-key:v1"));
		try {
			zeroizeAndClear(crypto);
			unlockedMasterKey = masterKey.slice();
		} finally {
			crypto.zeroize(masterKey);
		}
	} finally {
		crypto.zeroize(sessionUnwrapKey);
		crypto.zeroize(wrappedMasterKey);
	}
}

export async function prepareUpload(input) {
	const crypto = await ensureCrypto();
	if (!unlockedMasterKey) {
		throw new Error("Vault is locked");
	}
	const metadata = new TextEncoder().encode(JSON.stringify(input.metadata || {}));
	const fileKey = crypto.generate_file_key();
	try {
		const encryptedMetadata = crypto.encrypt_chunk(metadata, fileKey, aadBytes(input.metadataAad));
		try {
			const encryptedFileKey = crypto.wrap_file_key(fileKey, unlockedMasterKey, aadBytes(input.fileKeyAad));
			try {
				activeUploads.set(input.uploadToken, { fileKey: fileKey.slice() });
				return {
					encryptedMetadata: encodeBase64(encryptedMetadata),
					encryptedFileKey: encodeBase64(encryptedFileKey),
					totalParts: Number(input.totalParts || 0),
					encryptionVersion: 1,
				};
			} finally {
				crypto.zeroize(encryptedFileKey);
			}
		} finally {
			crypto.zeroize(encryptedMetadata);
		}
	} finally {
		crypto.zeroize(metadata);
		crypto.zeroize(fileKey);
	}
}

export async function encryptUploadPart(input) {
	const crypto = await ensureCrypto();
	const upload = activeUploads.get(String(input.uploadToken || ""));
	if (!upload || !upload.fileKey) {
		throw new Error("Upload context is missing");
	}
	const chunkBytes = toUint8Array(input.chunkBytes);
	try {
		const encryptedChunk = crypto.encrypt_chunk(chunkBytes, upload.fileKey, aadBytes(input.aad));
		try {
			const encryptedHash = crypto.hash_bytes_blake3(encryptedChunk);
			try {
				return {
					encryptedChunk: encryptedChunk.slice(),
					encryptedHash: encodeBase64(encryptedHash),
					encryptedSize: encryptedChunk.length,
				};
			} finally {
				crypto.zeroize(encryptedHash);
			}
		} finally {
			crypto.zeroize(encryptedChunk);
		}
	} finally {
		crypto.zeroize(chunkBytes);
	}
}

export async function hashUploadPayload(input) {
	const crypto = await ensureCrypto();
	const bytes = toUint8Array(input && input.bytes);
	try {
		const hash = crypto.hash_bytes_blake3(bytes);
		try {
			return encodeBase64(hash);
		} finally {
			crypto.zeroize(hash);
		}
	} finally {}
}

export async function finalizeUpload(input) {
	const crypto = await ensureCrypto();
	const upload = activeUploads.get(String(input.uploadToken || ""));
	if (!upload || !upload.fileKey) {
		throw new Error("Upload context is missing");
	}
	const manifest = new TextEncoder().encode(JSON.stringify(input.manifest || {}));
	let encryptedHash = new Uint8Array();
	try {
		const encryptedManifest = crypto.encrypt_chunk(manifest, upload.fileKey, aadBytes(input.manifestAad));
		try {
				const partHashes = input.chunkHashes || input.partHashes || [];
			let totalLength = 0;
			const decoded = [];
			for (let i = 0; i < partHashes.length; i++) {
				const hashBytes = decodeBase64(partHashes[i]);
				decoded.push(hashBytes);
				totalLength += hashBytes.length;
			}
			const combined = new Uint8Array(totalLength);
			let offset = 0;
			for (let i = 0; i < decoded.length; i++) {
				combined.set(decoded[i], offset);
				offset += decoded[i].length;
				crypto.zeroize(decoded[i]);
			}
			try {
				encryptedHash = crypto.hash_bytes_blake3(combined);
			} finally {
				crypto.zeroize(combined);
			}
			return {
				encryptedManifest: encodeBase64(encryptedManifest),
				encryptedHash: encodeBase64(encryptedHash),
			};
		} finally {
			crypto.zeroize(encryptedManifest);
		}
	} finally {
		crypto.zeroize(manifest);
		crypto.zeroize(encryptedHash);
		crypto.zeroize(upload.fileKey);
		activeUploads.delete(String(input.uploadToken || ""));
	}
}

export async function clearUploadContext(uploadToken) {
	const crypto = await ensureCrypto();
	const upload = activeUploads.get(String(uploadToken || ""));
	if (upload && upload.fileKey) {
		crypto.zeroize(upload.fileKey);
		activeUploads.delete(String(uploadToken || ""));
	}
}
