function renameState() {
  if (!window.__arkiveRenameState) {
    window.__arkiveRenameState = {
      activeEntry: null,
      activeInput: null,
      originalName: "",
      originalEditableName: "",
      lockedSuffix: "",
      submitting: false,
    };
  }
  return window.__arkiveRenameState;
}

function selectedEntries() {
  return window.ArkiveEntrySelection && typeof window.ArkiveEntrySelection.getSelectedEntries === "function"
    ? window.ArkiveEntrySelection.getSelectedEntries()
    : [];
}

function focusedEntry() {
  return window.ArkiveEntrySelection && typeof window.ArkiveEntrySelection.getFocusedEntry === "function"
    ? window.ArkiveEntrySelection.getFocusedEntry()
    : null;
}

function resolveRenameTarget() {
  const selected = selectedEntries();
  if (selected.length === 1 && window.ArkiveEntrySelection && typeof window.ArkiveEntrySelection.findEntryByID === "function") {
    return window.ArkiveEntrySelection.findEntryByID(selected[0].id);
  }
  return focusedEntry();
}

function nameField(entry) {
  return entry ? entry.querySelector("[data-file-field='name'], [data-folder-field='name']") : null;
}

function entryType(entry) {
  return String(entry && entry.getAttribute("data-entry-type") || "");
}

function entryID(entry) {
  return String(entry && entry.getAttribute("data-entry-id") || "");
}

function currentName(entry) {
  const type = entryType(entry);
  if (type === "folder") {
    return String(entry.getAttribute("data-folder-name") || "").trim();
  }
  return String(entry.getAttribute("data-file-name") || "").trim();
}

function splitEditableFileName(name) {
  const value = String(name || "");
  if (!value) {
    return { editableName: "", lockedSuffix: "" };
  }
  if (value[0] === "." && value.indexOf(".", 1) === -1) {
    return { editableName: value, lockedSuffix: "" };
  }
  const lastDot = value.lastIndexOf(".");
  if (lastDot <= 0 || lastDot === value.length - 1) {
    return { editableName: value, lockedSuffix: "" };
  }
  return {
    editableName: value.slice(0, lastDot),
    lockedSuffix: value.slice(lastDot),
  };
}

function parseJSONAttribute(entry, name) {
  const raw = String(entry && entry.getAttribute(name) || "");
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw);
  } catch (_) {
    return null;
  }
}

async function loadFileMetadata(entry) {
  const cached = parseJSONAttribute(entry, "data-file-metadata-json");
  const cachedVaultID = String(entry.getAttribute("data-file-vault-id") || "");
  if (cached && cachedVaultID) {
    return { metadata: cached, vaultId: cachedVaultID };
  }
  if (!window.ArkiveFileReader) {
    return null;
  }
  const reader = new window.ArkiveFileReader({ fileId: entryID(entry) });
  try {
    await reader.load();
    const metadata = reader.getMetadata() || {};
    const vaultId = String((reader.record && reader.record.vaultId) || "");
    try {
      entry.setAttribute("data-file-metadata-json", JSON.stringify(metadata));
    } catch (_) {}
    if (vaultId) {
      entry.setAttribute("data-file-vault-id", vaultId);
    }
    return { metadata: metadata, vaultId: vaultId };
  } catch (_) {
    return null;
  } finally {
    await reader.dispose();
  }
}

async function loadFolderMetadata(entry) {
  const cached = parseJSONAttribute(entry, "data-folder-metadata-json");
  if (cached) {
    return cached;
  }
  if (!window.ArkiveVault) {
    return null;
  }
  const encryptedMetadata = String(entry.getAttribute("data-folder-meta-b64") || "");
  const encryptedName = String(entry.getAttribute("data-folder-name-b64") || "");
  try {
    let metadata = null;
    if (encryptedMetadata && typeof window.ArkiveVault.decryptFolderMetadata === "function") {
      const result = await window.ArkiveVault.decryptFolderMetadata(encryptedMetadata);
      metadata = result && result.metadata ? result.metadata : null;
    }
    if ((!metadata || !metadata.name) && encryptedName && typeof window.ArkiveVault.decryptFolderName === "function") {
      const result = await window.ArkiveVault.decryptFolderName(encryptedName);
      metadata = result && result.metadata ? result.metadata : metadata;
    }
    if (!metadata) {
      return null;
    }
    try {
      entry.setAttribute("data-folder-metadata-json", JSON.stringify(metadata));
    } catch (_) {}
    return metadata;
  } catch (_) {
    return null;
  }
}

function updateFileName(entry, name, metadata, encryptedMetadataB64) {
  entry.querySelectorAll("[data-file-field='name']").forEach(function(node) {
    node.textContent = name;
    node.removeAttribute("aria-hidden");
  });
  entry.setAttribute("data-file-name", name);
  if (encryptedMetadataB64) {
    entry.setAttribute("data-file-metadata-b64", encryptedMetadataB64);
  }
  try {
    entry.setAttribute("data-file-metadata-json", JSON.stringify(metadata || {}));
  } catch (_) {}
  entry.classList.add("is-hydrated");

  document.querySelectorAll("[data-file-item='" + CSS.escape(entryID(entry)) + "']").forEach(function(node) {
    if (node === entry) {
      return;
    }
    node.querySelectorAll("[data-file-field='name']").forEach(function(nameNode) {
      nameNode.textContent = name;
      nameNode.removeAttribute("aria-hidden");
    });
    node.setAttribute("data-file-name", name);
    if (encryptedMetadataB64) {
      node.setAttribute("data-file-metadata-b64", encryptedMetadataB64);
    }
    try {
      node.setAttribute("data-file-metadata-json", JSON.stringify(metadata || {}));
    } catch (_) {}
    node.classList.add("is-hydrated");
  });
}

function updateFolderName(entry, name, metadata, encryptedNameB64, encryptedMetadataB64) {
  const id = entryID(entry);
  const selectors = [
    "[data-folder-item='" + CSS.escape(id) + "']",
    "[data-folder-breadcrumb='" + CSS.escape(id) + "']",
  ];
  document.querySelectorAll(selectors.join(",")).forEach(function(node) {
    node.querySelectorAll("[data-folder-field='name']").forEach(function(nameNode) {
      nameNode.textContent = name;
      nameNode.removeAttribute("aria-hidden");
    });
    node.setAttribute("data-folder-name", name);
    if (encryptedNameB64) {
      node.setAttribute("data-folder-name-b64", encryptedNameB64);
    }
    if (encryptedMetadataB64) {
      node.setAttribute("data-folder-meta-b64", encryptedMetadataB64);
    }
    try {
      node.setAttribute("data-folder-metadata-json", JSON.stringify(metadata || {}));
    } catch (_) {}
    node.classList.add("is-hydrated");
  });

  const uploadLabel = document.querySelector("[data-upload-label='true']");
  const currentFolder = document.querySelector("[data-current-folder-id]");
  if (uploadLabel && currentFolder && currentFolder.getAttribute("data-current-folder-id") === id) {
    uploadLabel.textContent = name ? "Upload to " + name : "Upload here";
  }
}

function finishRename(cancel) {
  const state = renameState();
  const entry = state.activeEntry;
  const input = state.activeInput;
  const field = nameField(entry);
  if (!entry || !input || !field) {
    state.activeEntry = null;
    state.activeInput = null;
    state.originalName = "";
    state.originalEditableName = "";
    state.lockedSuffix = "";
    state.submitting = false;
    return;
  }

  if (cancel) {
    field.textContent = state.originalName || " ";
    field.removeAttribute("hidden");
  } else {
    if (!field.textContent.trim()) {
      field.textContent = state.originalName || " ";
    }
    field.removeAttribute("hidden");
  }

  entry.classList.remove("is-renaming");
  if (input.parentNode) {
    input.parentNode.removeChild(input);
  }
  try {
    entry.focus();
  } catch (_) {}
  state.activeEntry = null;
  state.activeInput = null;
  state.originalName = "";
  state.originalEditableName = "";
  state.lockedSuffix = "";
  state.submitting = false;
}

async function submitRename() {
  const state = renameState();
  const entry = state.activeEntry;
  const input = state.activeInput;
  if (!entry || !input || state.submitting) {
    return;
  }
  const nextEditableName = String(input.value || "").trim();
  if (!nextEditableName) {
    finishRename(true);
    return;
  }
  const nextName = entryType(entry) === "file"
    ? nextEditableName + state.lockedSuffix
    : nextEditableName;
  if (nextName === state.originalName) {
    finishRename(true);
    return;
  }

  state.submitting = true;
  input.disabled = true;
  entry.classList.add("is-rename-saving");

  try {
    if (!window.ArkiveVault) {
      throw new Error("Vault is unavailable.");
    }

    const type = entryType(entry);
    if (type === "folder") {
      const current = await loadFolderMetadata(entry);
      const metadata = Object.assign({}, current || {}, { name: nextName });
      const encryptedName = await window.ArkiveVault.encryptFolderName({ name: nextName });
      const encryptedMetadata = await window.ArkiveVault.encryptFolderMetadata(metadata);
      const encryptedNameB64 = encryptedName && encryptedName.encryptedMetadata ? encryptedName.encryptedMetadata : "";
      const encryptedMetadataB64 = encryptedMetadata && encryptedMetadata.encryptedMetadata ? encryptedMetadata.encryptedMetadata : "";
      if (!encryptedNameB64 || !encryptedMetadataB64) {
        throw new Error("Rename failed.");
      }
      const response = await fetch("/api/entries/rename", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          type: "folder",
          id: entryID(entry),
          encryptedName: encryptedNameB64,
          encryptedMetadata: encryptedMetadataB64,
        })
      });
      if (!response.ok) {
        throw new Error("Rename failed.");
      }
      updateFolderName(entry, nextName, metadata, encryptedNameB64, encryptedMetadataB64);
    } else {
      if (!window.ArkiveFileReader || !window.ArkiveVault.encryptFileMetadataInContext) {
        throw new Error("Rename failed.");
      }
      const reader = new window.ArkiveFileReader({ fileId: entryID(entry) });
      let metadata = null;
      let encryptedMetadataB64 = "";
      try {
        await reader.load();
        metadata = Object.assign({}, reader.getMetadata() || {}, { name: nextName });
        const vaultId = String((reader.record && reader.record.vaultId) || "");
        const fileId = String((reader.record && reader.record.fileId) || entryID(entry));
        if (!vaultId || !fileId) {
          throw new Error("Rename failed.");
        }
        const encryptedMetadata = await window.ArkiveVault.encryptFileMetadataInContext(reader.contextId, metadata, vaultId, fileId);
        encryptedMetadataB64 = encryptedMetadata && encryptedMetadata.encryptedMetadata ? encryptedMetadata.encryptedMetadata : "";
        if (!encryptedMetadataB64) {
          throw new Error("Rename failed.");
        }
      } finally {
        await reader.dispose();
      }
      const response = await fetch("/api/entries/rename", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          type: "file",
          id: entryID(entry),
          encryptedName: "",
          encryptedMetadata: encryptedMetadataB64,
        })
      });
      if (!response.ok) {
        throw new Error("Rename failed.");
      }
      updateFileName(entry, nextName, metadata, encryptedMetadataB64);
    }

    finishRename(false);
    if (window.Toast) {
      window.Toast.success("Name updated.", { title: "Renamed" });
    }
  } catch (error) {
    entry.classList.remove("is-rename-saving");
    input.disabled = false;
    if (window.Toast) {
      window.Toast.error((error && error.message) || "Rename failed.");
    }
    finishRename(true);
  }
}

function bindRenameInput(input) {
  input.addEventListener("keydown", function(event) {
    if (event.key === "Enter") {
      event.preventDefault();
      event.stopPropagation();
      submitRename();
      return;
    }
    if (event.key === "Escape") {
      event.preventDefault();
      event.stopPropagation();
      finishRename(true);
    }
  });
  input.addEventListener("blur", function() {
    window.setTimeout(function() {
      const state = renameState();
      if (state.activeInput === input && !state.submitting) {
        finishRename(true);
      }
    }, 0);
  });
}

async function beginRename(entry) {
  if (!entry) {
    return;
  }
  const state = renameState();
  if (state.activeEntry && state.activeEntry !== entry) {
    finishRename(true);
  }
  if (state.activeEntry === entry) {
    return;
  }

  const field = nameField(entry);
  if (!field) {
    return;
  }

  const existingName = currentName(entry);
  if (!existingName) {
    if (entryType(entry) === "folder") {
      const metadata = await loadFolderMetadata(entry);
      if (metadata && metadata.name) {
        entry.setAttribute("data-folder-name", String(metadata.name));
      }
    } else {
      const fileState = await loadFileMetadata(entry);
      if (fileState && fileState.metadata && fileState.metadata.name) {
        entry.setAttribute("data-file-name", String(fileState.metadata.name));
      }
    }
  }

  const resolvedName = currentName(entry);
  if (!resolvedName) {
    if (window.Toast) {
      window.Toast.error("Rename unavailable until item metadata is ready.");
    }
    return;
  }

  const input = document.createElement("input");
  input.type = "text";
  input.className = "files-rename-input";
  input.setAttribute("aria-label", "Rename entry");

  let editableName = resolvedName;
  let lockedSuffix = "";
  if (entryType(entry) === "file") {
    const split = splitEditableFileName(resolvedName);
    editableName = split.editableName;
    lockedSuffix = split.lockedSuffix;
  }
  input.value = editableName;

  state.activeEntry = entry;
  state.activeInput = input;
  state.originalName = resolvedName;
  state.originalEditableName = editableName;
  state.lockedSuffix = lockedSuffix;
  state.submitting = false;

  entry.classList.add("is-renaming");
  field.setAttribute("hidden", "hidden");
  field.parentNode.insertBefore(input, field.nextSibling);
  bindRenameInput(input);
  input.focus();
  input.select();
}

function requestRename(entry) {
  const target = entry || resolveRenameTarget();
  if (!target) {
    return;
  }
  beginRename(target);
}

export function initRenameEntries() {
  window.ArkiveRenameEntries = {
    requestRename: requestRename,
    finishRename: finishRename,
    activeEntry: function() {
      return renameState().activeEntry;
    },
  };

  document.addEventListener("arkive:rename-request", function(event) {
    const detail = event && event.detail ? event.detail : {};
    requestRename(detail.entry || null);
  });
}
