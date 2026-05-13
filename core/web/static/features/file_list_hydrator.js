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

function updateItem(item, metadata, plaintextSize) {
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
  void metadata;
  void reader;
  // Grid cards stay metadata-only until Arkive has bounded thumbnail assets.
  clearPreview(previewEl);
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
    updateItem(item, metadata, reader.record ? reader.record.plaintextSize : 0);
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
  for (let i = 0; i < items.length; i++) {
    await hydrateItem(items[i]);
  }
}
