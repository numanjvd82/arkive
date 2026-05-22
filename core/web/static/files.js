import { setButtonBusy } from "./features/button_state.js";
import { Dialog } from "./features/dialog.js";
import { ArkiveFileReader } from "./features/file_reader.js";
import { entrySelection } from "./features/file_selection.js";
import { Toast } from "./features/toast.js";
import { vault, waitUntilReady } from "./features/vault.js";
import { apiRequest } from "./lib/api.js";
import { showAppError } from "./lib/toasts.js";
import { thumbnailCache } from "./upload/thumbnail_cache.js";

export const filesActions = {};

(function() {
  async function downloadFile(fileId, trigger) {
    if (!fileId) {
      return;
    }
    if (trigger) {
      trigger.disabled = true;
    }
    const reader = new ArkiveFileReader({ fileId: fileId });
    try {
      await reader.download();
    } catch (error) {
      showAppError(error, {
        code: "download_failed",
        message: "Download failed.",
      });
    } finally {
      if (trigger) {
        trigger.disabled = false;
      }
      await reader.dispose();
    }
  }

  filesActions.downloadFile = downloadFile;

  document.addEventListener("click", async function(event) {
    const target = event.target.closest("[data-file-action='download']");
    if (!target) {
      return;
    }
    event.preventDefault();
    await downloadFile(target.getAttribute("data-file-id"), target);
  });
})();

(function() {
  const fileRows = document.querySelectorAll("[data-file-row]");
  const fileCards = document.querySelectorAll("[data-file-card]");

  function openEntry(node) {
    if (!node) {
      return;
    }
    const href = node.getAttribute("data-file-open") || node.getAttribute("data-folder-open") || "";
    if (href) {
      window.location.href = href;
    }
  }

  function bindOpen(node) {
    if (!node || node.hasAttribute("data-file-open-bound")) {
      return;
    }
    node.setAttribute("data-file-open-bound", "true");
    node.addEventListener("dblclick", function(event) {
      if (event.target.closest("a, button, input, label")) {
        return;
      }
      openEntry(node);
    });
    node.addEventListener("keydown", function(event) {
      if (event.key !== "Enter") {
        return;
      }
      openEntry(node);
    });
  }

  filesActions.openEntry = openEntry;
  fileRows.forEach(bindOpen);
  fileCards.forEach(bindOpen);
})();

(function() {
  const DELETE_TOAST_STORAGE_KEY = "arkive:entries-delete-toast:v1";
  const backdrop = document.getElementById("file-delete-backdrop");
  const title = document.getElementById("file-delete-title");
  const meta = document.getElementById("file-delete-meta");
  const cancelButton = document.getElementById("file-delete-cancel");
  const confirmButton = document.getElementById("file-delete-confirm");
  let pendingDeleteEntries = [];

  if (!confirmButton) {
    return;
  }

  function fileName(fileId) {
    const item = document.querySelector("[data-file-item='" + fileId + "']");
    return item ? String(item.getAttribute("data-file-name") || "") : "";
  }

  function folderName(folderId) {
    const item = document.querySelector("[data-folder-item='" + folderId + "']");
    return item ? String(item.getAttribute("data-folder-name") || "") : "";
  }

  function selectedFileNames(ids) {
    return ids.map(fileName).filter(Boolean);
  }

  function selectedFolderNames(ids) {
    return ids.map(folderName).filter(Boolean);
  }

  function removeEntries(entries) {
    entries.forEach(function(entry) {
      const selector = entry.type === "folder" ? "[data-folder-item='" + entry.id + "']" : "[data-file-item='" + entry.id + "']";
      document.querySelectorAll(selector).forEach(function(item) {
        if (item && item.parentNode) {
          item.parentNode.removeChild(item);
        }
      });
    });
    entrySelection.clear();
  }

  function pageHasRows() {
    return document.querySelector("[data-file-item], [data-folder-item]") !== null;
  }

  function clearThumbnailCache(fileIds) {
    if (!thumbnailCache || typeof thumbnailCache.deleteForFiles !== "function") {
      return Promise.resolve();
    }
    return thumbnailCache.deleteForFiles(fileIds);
  }

  function describeDeleteResult(result) {
    const fileCount = Number(result && result.deletedFiles || 0);
    const folderCount = Number(result && result.deletedFolders || 0);
    if (fileCount > 0 && folderCount > 0) {
      return "Deleted " + folderCount + " folder" + (folderCount === 1 ? "" : "s") + " and " + fileCount + " file" + (fileCount === 1 ? "" : "s") + ".";
    }
    if (folderCount > 0) {
      return "Deleted " + folderCount + " folder" + (folderCount === 1 ? "" : "s") + ".";
    }
    if (fileCount > 0) {
      return "Deleted " + fileCount + " file" + (fileCount === 1 ? "" : "s") + ".";
    }
    return "Deleted items.";
  }

  function storeDeleteToast(message) {
    try {
      window.sessionStorage.setItem(DELETE_TOAST_STORAGE_KEY, message);
    } catch (_) {}
  }

  function flushStoredDeleteToast() {
    let message = "";
    try {
      message = String(window.sessionStorage.getItem(DELETE_TOAST_STORAGE_KEY) || "");
      if (message) {
        window.sessionStorage.removeItem(DELETE_TOAST_STORAGE_KEY);
      }
    } catch (_) {}
    if (message) {
      Toast.success(message, { title: "Deleted" });
    }
  }

  function describeDelete(entries) {
    const fileEntries = entries.filter(function(entry) { return entry && entry.type === "file"; });
    const folderEntries = entries.filter(function(entry) { return entry && entry.type === "folder"; });
    const fileCount = fileEntries.length;
    const folderCount = folderEntries.length;

    if (title) {
      title.textContent = fileCount+folderCount === 1 ? "Delete item?" : "Delete items?";
    }
    if (!meta) {
      return;
    }

    if (folderCount === 0 && fileCount === 1) {
      const names = selectedFileNames([fileEntries[0].id]);
      meta.textContent = names[0]
        ? "This will permanently delete " + names[0] + ". This action cannot be undone."
        : "This will permanently delete the file. This action cannot be undone.";
      return;
    }

    if (fileCount === 0 && folderCount === 1) {
      const names = selectedFolderNames([folderEntries[0].id]);
      meta.textContent = names[0]
        ? "This will permanently delete " + names[0] + " and everything inside it. This action cannot be undone."
        : "This will permanently delete the folder and everything inside it. This action cannot be undone.";
      return;
    }

    if (folderCount === 0) {
      meta.textContent = "This will permanently delete " + fileCount + " file" + (fileCount === 1 ? "" : "s") + ". This action cannot be undone.";
      return;
    }

    if (fileCount === 0) {
      meta.textContent = "This will permanently delete " + folderCount + " folder" + (folderCount === 1 ? "" : "s") + " and everything inside " + (folderCount === 1 ? "it." : "them.") + " This action cannot be undone.";
      return;
    }

    meta.textContent = "This will permanently delete " + fileCount + " file" + (fileCount === 1 ? "" : "s") + ", " + folderCount + " folder" + (folderCount === 1 ? "" : "s") + ", and everything inside those folders. This action cannot be undone.";
  }

  function openDialog(entries) {
    if (!backdrop || !confirmButton) {
      return;
    }
    pendingDeleteEntries = entries.slice();
    describeDelete(pendingDeleteEntries);
    Dialog.open("file-delete-backdrop");
    setButtonBusy(confirmButton, false);
  }

  function closeDialog() {
    if (!backdrop) {
      return;
    }
    pendingDeleteEntries = [];
    Dialog.close("file-delete-backdrop");
  }

  flushStoredDeleteToast();

  filesActions.requestDeleteFiles = function(fileIds) {
    const ids = Array.isArray(fileIds) ? fileIds.filter(Boolean) : [];
    if (!ids.length) {
      return;
    }
    openDialog(ids.map(function(id) {
      return { id: id, type: "file" };
    }));
  };

  filesActions.requestDeleteEntries = function(entries) {
    const normalized = Array.isArray(entries) ? entries.filter(function(entry) {
      return entry && entry.id && (entry.type === "file" || entry.type === "folder");
    }) : [];
    if (!normalized.length) {
      return;
    }
    openDialog(normalized);
  };

  document.addEventListener("click", function(event) {
    const button = event.target.closest("[data-file-action='delete']");
    if (!button) {
      return;
    }
    const fileId = button.getAttribute("data-file-id");
    if (!fileId) {
      return;
    }
    openDialog([{ id: fileId, type: "file" }]);
  });

  if (cancelButton) {
    cancelButton.addEventListener("click", function() {
      closeDialog();
    });
  }

  if (confirmButton) {
    confirmButton.addEventListener("click", async function() {
      if (!pendingDeleteEntries.length) {
        closeDialog();
        return;
      }
      setButtonBusy(confirmButton, true, { busyText: "Deleting..." });
      const entries = pendingDeleteEntries.slice();
      const fileIds = entries.filter(function(entry) {
        return entry.type === "file";
      }).map(function(entry) {
        return entry.id;
      });
      const folderIds = entries.filter(function(entry) {
        return entry.type === "folder";
      }).map(function(entry) {
        return entry.id;
      });

      try {
        const result = await apiRequest("/api/entries/delete", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ fileIds: fileIds, folderIds: folderIds })
        }, {
          code: "conflict",
          message: "Delete failed",
        });
        const successMessage = describeDeleteResult(result);
        await clearThumbnailCache(fileIds).catch(function() {});
        removeEntries(entries);
        closeDialog();
        if (!pageHasRows() || folderIds.length > 0 || entries.length > 1) {
          storeDeleteToast(successMessage);
          window.location.reload();
          return;
        }
        Toast.success(successMessage, { title: "Deleted" });
      } catch (error) {
        showAppError(error, {
          code: "conflict",
          message: "Delete failed. Try again.",
        });
        closeDialog();
      } finally {
        setButtonBusy(confirmButton, false);
      }
    });
  }

  document.addEventListener("arkive:entries-delete-request", function(event) {
    const detail = event && event.detail ? event.detail : {};
    const selectedEntries = Array.isArray(detail.selectedEntries) ? detail.selectedEntries : [];
    if (!selectedEntries.length) {
      return;
    }
    openDialog(selectedEntries);
  });
})();

(function() {
  const backdrop = document.getElementById("file-share-backdrop");
  const linkInput = document.getElementById("share-link-input");
  const copyButton = document.getElementById("share-copy-button");
  const fileNameEl = document.getElementById("share-file-name");
  const statusEl = document.getElementById("share-status");
  const saveStateEl = document.getElementById("share-save-state");
  const stateNoteEl = document.getElementById("share-state-note");
  const saveButton = document.getElementById("share-save-button");
  const expirySelect = document.getElementById("share-expiry-select");
  const expiryToggle = document.getElementById("share-expiry-toggle");
  const expiryCustomWrap = document.querySelector(".share-expiry-custom");
  const expiryCustomInput = document.getElementById("share-expiry-custom");
  const expiryTimeInput = document.getElementById("share-expiry-time");
  const passwordToggle = document.getElementById("share-password-toggle");
  const shareModeStable = document.getElementById("share-mode-stable");
  const shareModeOnce = document.getElementById("share-mode-once");
  const passwordField = document.querySelector(".share-password-field");
  const passwordInput = document.getElementById("share-password");
  const passwordHelper = document.getElementById("share-password-helper");
  const passwordStrength = document.getElementById("share-password-strength");
  const revokeButton = document.getElementById("share-revoke-button");
  const deleteButton = document.getElementById("share-delete-button");
  const closeButton = document.getElementById("share-close-button");
  const errorEl = document.getElementById("share-error");

  let activeFileId = null;
  let activeShareId = "";
  let activeShareSecret = "";
  let activeShareStatus = "";

  if (!backdrop || !saveButton || !statusEl) {
    return;
  }

  function setShareError(message) {
    if (!errorEl) {
      return;
    }
    errorEl.textContent = message || "";
    errorEl.classList.toggle("is-visible", !!message);
  }

  function setSaveState(text, state) {
    if (!saveStateEl) {
      return;
    }
    saveStateEl.textContent = text || "";
    saveStateEl.classList.toggle("is-empty", !text);
    saveStateEl.classList.toggle("is-saving", state === "saving");
    saveStateEl.classList.toggle("is-error", state === "error");
  }

  function setButtonLabel(button, text) {
    if (!button) {
      return;
    }
    const label = button.querySelector(".button-label");
    if (label) {
      label.textContent = text;
      return;
    }
    button.textContent = text;
  }

  function setStateNote(message) {
    if (!stateNoteEl) {
      return;
    }
    const text = String(message || "");
    stateNoteEl.textContent = text;
    stateNoteEl.classList.toggle("is-visible", !!text);
  }

  function setDisabled(node, disabled) {
    if (!node) {
      return;
    }
    node.disabled = !!disabled;
  }

  function setShareEditable(editable) {
    setDisabled(passwordToggle, !editable);
    setDisabled(passwordInput, !editable || !(passwordToggle && passwordToggle.checked));
    setDisabled(expiryToggle, !editable);
    setDisabled(expirySelect, !editable || !(expiryToggle && expiryToggle.checked));
    setDisabled(expiryCustomInput, !editable || !(expiryToggle && expiryToggle.checked));
    setDisabled(expiryTimeInput, !editable || !(expiryToggle && expiryToggle.checked));
    setDisabled(shareModeStable, !editable);
    setDisabled(shareModeOnce, !editable);
    if (saveButton) {
      saveButton.hidden = !editable;
      saveButton.disabled = !editable;
    }
  }

  function setPasswordVisible(visible) {
    if (passwordField) {
      passwordField.classList.toggle("is-visible", !!visible);
    }
    if (passwordHelper) {
      passwordHelper.classList.toggle("is-visible", !!visible);
    }
    if (!visible && passwordStrength) {
      passwordStrength.textContent = "";
      passwordStrength.classList.remove("is-error");
      passwordStrength.classList.remove("is-success");
    }
    setDisabled(passwordInput, !visible || !passwordToggle || passwordToggle.disabled || !passwordToggle.checked);
  }

  function setExpiryVisible(visible) {
    if (expiryCustomWrap) {
      expiryCustomWrap.classList.toggle("is-visible", !!visible);
    }
    setDisabled(expirySelect, !visible || !expiryToggle || expiryToggle.disabled || !expiryToggle.checked);
    setDisabled(expiryCustomInput, !visible || !expiryToggle || expiryToggle.disabled || !expiryToggle.checked);
    setDisabled(expiryTimeInput, !visible || !expiryToggle || expiryToggle.disabled || !expiryToggle.checked);
  }

  function normalizeShareLink(link) {
    return String(link || "").trim();
  }

  function currentExpiryISO() {
    const dateValue = expiryCustomInput ? String(expiryCustomInput.value || "") : "";
    const timeValue = expiryTimeInput ? String(expiryTimeInput.value || "") : "";
    if (!dateValue) {
      return "";
    }
    return timeValue ? dateValue + "T" + timeValue + ":00" : dateValue + "T00:00:00";
  }

  function applyShareState(data) {
    const link = normalizeShareLink(data && data.url);
    const hasPassword = !!(data && data.hasPassword);
    const burnAfterRead = !!(data && data.burnAfterRead);
    const expiresAt = data && data.expiresAt ? String(data.expiresAt) : "";
    const shareStatus = data && data.status ? String(data.status) : "";
    activeShareId = String((data && data.id) || "");
    activeShareStatus = shareStatus;

    if (linkInput) {
      linkInput.value = link;
    }
    if (statusEl) {
      if (shareStatus === "revoked") {
        statusEl.textContent = "Revoked";
      } else if (shareStatus === "expired") {
        statusEl.textContent = "Expired";
      } else if (shareStatus === "burned") {
        statusEl.textContent = "Burned";
      } else {
        statusEl.textContent = link ? "Live" : "Ready to create";
      }
    }
    if (copyButton) {
      copyButton.disabled = !link;
    }
    if (deleteButton) {
      deleteButton.disabled = !activeShareId;
    }
    if (revokeButton) {
      revokeButton.hidden = !activeShareId || shareStatus === "burned";
      revokeButton.disabled = !activeShareId || shareStatus === "expired";
      setButtonLabel(revokeButton, shareStatus === "revoked" ? "Restore Link" : "Revoke Link");
    }
    if (passwordToggle) {
      passwordToggle.checked = hasPassword;
    }
    if (shareModeStable) {
      shareModeStable.checked = !burnAfterRead;
    }
    if (shareModeOnce) {
      shareModeOnce.checked = burnAfterRead;
    }
    if (expiryToggle) {
      expiryToggle.checked = !!expiresAt;
    }
    setPasswordVisible(hasPassword);
    setExpiryVisible(!!expiresAt);
    updatePasswordHelper(hasPassword);
    updatePasswordStrength("");

    if (!activeShareId) {
      setShareEditable(true);
      setStateNote("");
    } else if (shareStatus === "revoked") {
      setShareEditable(false);
      setStateNote("This link is revoked. Restore it before editing settings or using it again.");
    } else if (shareStatus === "burned") {
      setShareEditable(false);
      setStateNote("This one-time link has already been used and cannot be reused.");
    } else if (shareStatus === "expired") {
      setShareEditable(false);
      setStateNote("This link has expired. Delete it and create a new one if you still need to share the file.");
    } else {
      setShareEditable(true);
      setStateNote(burnAfterRead ? "This one-time link stays editable until it is used once." : "");
    }

    if (expiresAt && expiryCustomInput) {
      const parts = expiresAt.split("T");
      expiryCustomInput.value = parts[0] || "";
      if (expiryTimeInput) {
        expiryTimeInput.value = parts[1] ? parts[1].slice(0, 5) : "";
      }
      if (expirySelect) {
        expirySelect.value = "custom";
      }
    } else {
      if (expiryCustomInput) {
        expiryCustomInput.value = "";
      }
      if (expiryTimeInput) {
        expiryTimeInput.value = "";
      }
      if (expirySelect) {
        expirySelect.value = "custom";
      }
    }
  }

  function resetTransientFields() {
    if (passwordInput) {
      passwordInput.value = "";
    }
    activeShareId = "";
    activeShareSecret = "";
    activeShareStatus = "";
    setShareError("");
    setStateNote("");
    setSaveState("", "");
    updatePasswordStrength("");
  }

  function withAbsoluteURL(data) {
    if (!data || !data.token || (data.status && data.status !== "active")) {
      return data;
    }
    const shareSecret = String((data && data.shareSecret) || activeShareSecret || "");
    const hash = shareSecret ? "#s=" + encodeURIComponent(shareSecret) : "";
    return Object.assign({}, data, {
      url: window.location.origin + "/s/" + data.token + hash
    });
  }

  function loadFileRecord(fileId) {
    return apiRequest("/api/files/" + encodeURIComponent(fileId) + "/record", {
      method: "GET",
      headers: { "Content-Type": "application/json" }
    }, {
      code: "not_found",
      message: "Failed to load file",
    });
  }

  function passwordValidationMessage(password) {
    const value = String(password || "");
    if (!value) {
      return "Password is required.";
    }
    if (value.length < 8) {
      return "Use at least 8 characters.";
    }
    if (!/[a-z]/.test(value)) {
      return "Add a lowercase letter.";
    }
    if (!/[A-Z]/.test(value)) {
      return "Add an uppercase letter.";
    }
    if (!/[^A-Za-z0-9]/.test(value)) {
      return "Add a symbol.";
    }
    return "";
  }

  function updatePasswordHelper(hasExistingPassword) {
    if (!passwordHelper) {
      return;
    }
    passwordHelper.textContent = hasExistingPassword
      ? "Leave this blank to keep the current password, or enter a new one to rotate it."
      : "Use at least 8 characters with lowercase, uppercase, and a symbol.";
  }

  function updatePasswordStrength(message) {
    if (!passwordStrength) {
      return;
    }
    const text = String(message || "");
    passwordStrength.textContent = text;
    passwordStrength.classList.toggle("is-error", !!text);
    passwordStrength.classList.toggle("is-success", !text && !!(passwordToggle && passwordToggle.checked && passwordInput && passwordInput.value));
    if (!text && passwordToggle && passwordToggle.checked && passwordInput && passwordInput.value) {
      passwordStrength.textContent = "Strong enough for Arkive share protection.";
    }
  }

  function createRandomShareToken() {
    const bytes = new Uint8Array(24);
    window.crypto.getRandomValues(bytes);
    let binary = "";
    for (let i = 0; i < bytes.length; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
  }

  function createSharePayload(fileId, token) {
    return waitUntilReady()
      .then(function() {
        return loadFileRecord(fileId);
      })
      .then(function(record) {
        return vault.prepareShare(record, token).then(function(prepared) {
          return {
            encryptedShareKey: String((prepared && prepared.encryptedShareKey) || ""),
            encryptedFileKeyForShare: String((prepared && prepared.encryptedFileKeyForShare) || ""),
            shareSecret: String((prepared && prepared.shareSecret) || ""),
          };
        });
      });
  }

  function openShareDialog(fileId, fileName) {
    activeFileId = fileId;
    if (fileNameEl) {
      fileNameEl.textContent = fileName || "Encrypted file";
    }
    resetTransientFields();
    applyShareState(null);
    Dialog.open("file-share-backdrop");
    apiRequest("/api/files/" + encodeURIComponent(fileId) + "/share", {
      method: "GET",
      headers: { "Content-Type": "application/json" }
    }, {
      code: "not_found",
      message: "Failed to load share settings.",
    })
      .catch(function(error) {
        if (error && error.status === 404) {
          return null;
        }
        throw error;
      })
      .then(function(data) {
        if (!data) {
          return;
        }
        if (data && data.encryptedShareKey) {
          return waitUntilReady()
            .then(function() {
              return vault.openShareKey(data.encryptedShareKey, data.token || "");
            })
            .then(function(result) {
              activeShareSecret = String((result && result.shareSecret) || "");
              return data;
            });
        }
        activeShareSecret = "";
        return data;
      })
      .then(function(data) {
        if (!data) {
          applyShareState(null);
          return;
        }
        applyShareState(withAbsoluteURL(data));
      })
      .catch(function(error) {
        setShareError((error && error.message) || "Failed to load share settings.");
      });
  }

  filesActions.openShare = openShareDialog;

  function closeShareDialog() {
    activeFileId = null;
    Dialog.close("file-share-backdrop");
  }

  document.addEventListener("click", function(event) {
    const button = event.target.closest("[data-file-action='share']");
    if (!button) {
      return;
    }
    const fileId = button.getAttribute("data-file-id") || "";
    if (!fileId) {
      return;
    }
    openShareDialog(fileId, button.getAttribute("data-file-name") || "");
  });

  if (copyButton) {
    copyButton.addEventListener("click", async function() {
      if (copyButton.disabled) {
        return;
      }
      const value = linkInput ? String(linkInput.value || "") : "";
      if (!value) {
        return;
      }
      try {
        await navigator.clipboard.writeText(value);
        Toast.success("Share link copied.", { title: "Copied" });
      } catch (_) {
        showAppError(null, {
          code: "unknown_error",
          message: "Copy failed.",
        });
      }
    });
  }

  if (passwordToggle) {
    passwordToggle.addEventListener("change", function() {
      setPasswordVisible(passwordToggle.checked);
      updatePasswordHelper(!!activeShareId && passwordToggle.checked);
      updatePasswordStrength(passwordToggle.checked ? passwordValidationMessage(passwordInput && passwordInput.value) : "");
    });
  }

  if (passwordInput) {
    passwordInput.addEventListener("input", function() {
      if (!passwordToggle || !passwordToggle.checked) {
        return;
      }
      const value = String(passwordInput.value || "");
      if (!value && activeShareId) {
        updatePasswordStrength("");
        return;
      }
      updatePasswordStrength(passwordValidationMessage(value));
    });
  }

  if (expiryToggle) {
    expiryToggle.addEventListener("change", function() {
      setExpiryVisible(expiryToggle.checked);
    });
  }

  if (expirySelect) {
    expirySelect.addEventListener("change", function() {
      const value = String(expirySelect.value || "custom");
      if (value === "custom") {
        setExpiryVisible(expiryToggle && expiryToggle.checked);
        return;
      }
      const now = new Date();
      const hours = value === "1d" ? 24 : value === "7d" ? 24 * 7 : 24 * 30;
      now.setHours(now.getHours() + hours);
      if (expiryCustomInput) {
        expiryCustomInput.value = now.toISOString().slice(0, 10);
      }
      if (expiryTimeInput) {
        expiryTimeInput.value = now.toISOString().slice(11, 16);
      }
      setExpiryVisible(true);
      if (expiryToggle) {
        expiryToggle.checked = true;
      }
    });
  }

  if (saveButton) {
    saveButton.addEventListener("click", function() {
      if (!activeFileId) {
        return;
      }
      setButtonBusy(saveButton, true, { busyText: "Saving..." });
      setShareError("");
      setSaveState("Saving...", "saving");

      const payload = {
        password: passwordToggle && passwordToggle.checked ? String((passwordInput && passwordInput.value) || "") : "",
        expiresAt: expiryToggle && expiryToggle.checked ? currentExpiryISO() : "",
        requirePassword: !!(passwordToggle && passwordToggle.checked),
        burnAfterRead: !!(shareModeOnce && shareModeOnce.checked),
      };

      if (payload.requirePassword && (!activeShareId || payload.password)) {
        const passwordMessage = passwordValidationMessage(payload.password);
        if (passwordMessage) {
          setButtonBusy(saveButton, false);
          setSaveState("Save failed", "error");
          setShareError(passwordMessage);
          updatePasswordStrength(passwordMessage);
          return;
        }
      }

      const request = activeShareId
        ? apiRequest("/api/shares/" + encodeURIComponent(activeShareId), {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
          }, {
            code: "validation_failed",
            message: "Failed to save share settings.",
          })
        : (function() {
            const token = createRandomShareToken();
            return createSharePayload(activeFileId, token).then(function(cryptoPayload) {
            activeShareSecret = String((cryptoPayload && cryptoPayload.shareSecret) || "");
            return apiRequest("/api/files/" + encodeURIComponent(activeFileId) + "/share", {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({
                token: token,
                password: payload.password,
                expiresAt: payload.expiresAt,
                burnAfterRead: payload.burnAfterRead,
                encryptedShareKey: cryptoPayload.encryptedShareKey,
                encryptedFileKeyForShare: cryptoPayload.encryptedFileKeyForShare,
              }),
            }, {
              code: "validation_failed",
              message: "Failed to save share settings.",
            });
            });
          })();

      request
        .then(function(data) {
          activeShareId = String((data && data.id) || activeShareId || "");
          applyShareState(withAbsoluteURL(data));
          if (passwordInput) {
            passwordInput.value = "";
          }
          setSaveState("Saved", "");
        })
        .catch(function(error) {
          setSaveState("Save failed", "error");
          setShareError((error && error.message) || "Failed to save share settings.");
        })
        .finally(function() {
          setButtonBusy(saveButton, false);
        });
    });
  }

  if (deleteButton) {
    deleteButton.addEventListener("click", function() {
      if (!activeShareId) {
        return;
      }
      setButtonBusy(deleteButton, true, { busyText: "Deleting...", restoreDisabled: false });
      setShareError("");
      setSaveState("Deleting...", "saving");
      apiRequest("/api/shares/" + encodeURIComponent(activeShareId), {
        method: "DELETE",
        headers: { "Content-Type": "application/json" }
      }, {
        code: "unknown_error",
        message: "Failed to delete share",
      })
        .then(function() {
          applyShareState(null);
          activeShareSecret = "";
          setSaveState("Deleted", "");
        })
        .catch(function(error) {
          setSaveState("Delete failed", "error");
          setShareError((error && error.message) || "Failed to delete share.");
        })
        .finally(function() {
          setButtonBusy(deleteButton, false, { restoreDisabled: false });
        });
    });
  }

  if (revokeButton) {
    revokeButton.addEventListener("click", function() {
      if (!activeShareId) {
        return;
      }
      setButtonBusy(revokeButton, true, { busyText: "Revoking...", restoreDisabled: false });
      setShareError("");
      setSaveState("Revoking...", "saving");
      const isRevoked = activeShareStatus === "revoked";
      apiRequest("/api/shares/" + encodeURIComponent(activeShareId) + (isRevoked ? "/activate" : "/revoke"), {
        method: "POST",
        headers: { "Content-Type": "application/json" }
      }, {
        code: "unknown_error",
        message: isRevoked ? "Failed to restore share" : "Failed to revoke share",
      })
        .then(function(data) {
          setButtonBusy(revokeButton, false, { restoreDisabled: false });
          applyShareState(withAbsoluteURL(data));
          setSaveState(isRevoked ? "Restored" : "Revoked", "");
        })
        .catch(function(error) {
          setButtonBusy(revokeButton, false, { restoreDisabled: false });
          setSaveState(isRevoked ? "Restore failed" : "Revoke failed", "error");
          setShareError((error && error.message) || (isRevoked ? "Failed to restore share." : "Failed to revoke share."));
        });
    });
  }

  if (closeButton) {
    closeButton.addEventListener("click", function() {
      closeShareDialog();
    });
  }
})();
