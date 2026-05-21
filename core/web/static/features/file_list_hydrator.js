function formatBytes(bytes) {
  const value = Number(bytes || 0);
  if (!value) {
    return "0 B";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  const index = Math.min(
    units.length - 1,
    Math.floor(Math.log(value) / Math.log(1024)),
  );
  const amount = value / Math.pow(1024, index);
  return amount.toFixed(index > 0 ? 1 : 0) + " " + units[index];
}

const THUMBNAIL_HYDRATION_CONCURRENCY = 6;

function updateItem(item, metadata, plaintextSize, reader) {
  const nameEl = item.querySelector("[data-file-field='name']");
  const typeEl = item.querySelector("[data-file-field='type']");
  const sizeEl = item.querySelector("[data-file-field='size']");
  const shareEl = item.querySelector("[data-file-action='share']");
  const deleteEl = item.querySelector("[data-file-action='delete']");
  const shareDeleteEl = item.querySelector("[data-share-action='delete']");
  const realName = metadata && metadata.name ? metadata.name : "";
  const realType = metadata && metadata.mime ? metadata.mime : "";
  const realSize = metadata && metadata.size ? metadata.size : plaintextSize;

  if (nameEl) {
    nameEl.textContent = realName;
    nameEl.removeAttribute("aria-hidden");
  }
  if (typeEl) {
    typeEl.textContent = realType;
    typeEl.removeAttribute("aria-hidden");
    if (realType) {
      typeEl.setAttribute("title", realType);
    } else {
      typeEl.removeAttribute("title");
    }
  }
  if (sizeEl) {
    sizeEl.textContent = formatBytes(realSize);
  }
  if (shareEl) {
    shareEl.setAttribute("data-file-name", realName);
  }
  if (deleteEl) {
    deleteEl.setAttribute("data-file-name", realName);
  }
  if (shareDeleteEl) {
    shareDeleteEl.setAttribute("data-share-file", realName);
  }
  item.setAttribute("data-file-name", realName);
  if (reader && reader.record && reader.record.vaultId) {
    item.setAttribute("data-file-vault-id", String(reader.record.vaultId));
  }
  try {
    item.setAttribute("data-file-metadata-json", JSON.stringify(metadata || {}));
  } catch (_) {}

  item.classList.add("is-hydrated");
  item.removeAttribute("aria-busy");
}

function clearPreview(previewEl) {
  if (!previewEl) {
    return;
  }
  previewEl.classList.remove("has-media");
  const stale = previewEl.querySelector("[data-file-preview-media='true']");
  if (stale) {
    const objectURL = stale.getAttribute("data-object-url");
    if (objectURL) {
      URL.revokeObjectURL(objectURL);
    }
    if (stale.tagName === "VIDEO") {
      stale.pause();
      stale.removeAttribute("src");
      stale.load();
    }
    stale.parentNode.removeChild(stale);
  }
  const icon = previewEl.querySelector("[data-file-field='icon']");
  if (icon) {
    icon.hidden = false;
  }
}

async function updateGridPreview(card, metadata, reader) {
  const previewEl = card.querySelector("[data-file-preview='true']");
  if (!previewEl) {
    return;
  }
  clearPreview(previewEl);
  const preview = metadata && metadata.preview ? metadata.preview : null;
  if (!preview || !preview.has_thumbnail || !reader || typeof reader.readThumbnail !== "function") {
    return;
  }
  try {
    const thumbnailBytes = await reader.readThumbnail();
    const mime = String(preview.thumbnail_mime || "image/webp");
    const objectURL = URL.createObjectURL(new Blob([thumbnailBytes], { type: mime }));
    const image = document.createElement("img");
    image.className = "files-card-preview-media";
    image.setAttribute("data-file-preview-media", "true");
    image.setAttribute("data-object-url", objectURL);
    image.alt = metadata && metadata.name ? metadata.name : "Encrypted file thumbnail";
    image.loading = "lazy";
    image.decoding = "async";
    image.src = objectURL;
    previewEl.appendChild(image);
    previewEl.classList.add("has-media");
    const icon = previewEl.querySelector("[data-file-field='icon']");
    if (icon) {
      icon.hidden = true;
    }
  } catch (_) {
    clearPreview(previewEl);
  }
}

async function hydrateItem(item) {
  const fileId = item.getAttribute("data-file-item");
  if (!fileId) {
    return;
  }
  const reader = new window.ArkiveFileReader({ fileId: fileId });
  try {
    await reader.load();
    const metadata = reader.getMetadata();
    updateItem(item, metadata, reader.record ? reader.record.plaintextSize : 0, reader);
    if (item.hasAttribute("data-file-card")) {
      await updateGridPreview(item, metadata, reader);
    }
  } catch (_) {
  } finally {
    await reader.dispose();
  }
}

export async function initFileListHydrator() {
  const items = document.querySelectorAll("[data-file-item]");
  if (!items.length || !window.ArkiveFileReader || !window.ArkiveVault) {
    return;
  }
  if (typeof window.ArkiveVault.waitUntilReady === "function") {
    await window.ArkiveVault.waitUntilReady();
  }

  let nextIndex = 0;

  async function worker() {
    while (nextIndex < items.length) {
      const currentIndex = nextIndex++;
      await hydrateItem(items[currentIndex]);
    }
  }

  const workers = [];
  const count = Math.min(THUMBNAIL_HYDRATION_CONCURRENCY, items.length);
  for (let i = 0; i < count; i++) {
    workers.push(worker());
  }
  await Promise.all(workers);
}
