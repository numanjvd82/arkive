(function() {
  document.addEventListener("click", async function(event) {
    const target = event.target.closest("[data-file-action='download']");
    if (!target) {
      return;
    }
    event.preventDefault();
    const fileId = target.getAttribute("data-file-id");
    if (!fileId || !window.ArkiveFileReader) {
      return;
    }
    target.disabled = true;
    const reader = new window.ArkiveFileReader({ fileId: fileId });
    try {
      await reader.download();
    } catch (error) {
      if (window.Toast) {
        window.Toast.error((error && error.message) || "Download failed.");
      }
    } finally {
      target.disabled = false;
      await reader.dispose();
    }
  });
})();

(function() {
  const tableRows = Array.from(document.querySelectorAll("[data-file-row]"));
  const deleteButtons = document.querySelectorAll("[data-file-action='delete']");
  const selectAll = document.getElementById("files-select-all");
  const bulkDeleteButton = document.getElementById("files-delete-selected");
  const selectionCount = document.getElementById("files-selection-count");
  const selectionToolbar = document.getElementById("files-selection-toolbar");
  const gridCards = Array.from(document.querySelectorAll("[data-file-grid-select]"));
  const gridWrap = document.querySelector(".files-grid-wrap");
  const contextMenu = document.getElementById("files-grid-context-menu");
  const fileContextItems = Array.from(document.querySelectorAll("[data-grid-menu-action]"));
  const spaceContextItems = Array.from(document.querySelectorAll("[data-grid-space-action]"));
  const backdrop = document.getElementById("file-delete-backdrop");
  const meta = document.getElementById("file-delete-meta");
  const cancelButton = document.getElementById("file-delete-cancel");
  const confirmButton = document.getElementById("file-delete-confirm");
  let pendingDeleteIds = [];
  let activeContextFileId = "";

  if (!deleteButtons.length && !bulkDeleteButton && !gridCards.length && !tableRows.length) {
    return;
  }

  tableRows.forEach(function(row) {
    row.addEventListener("click", function(event) {
      if (event.target.closest("a, button, input, label")) {
        return;
      }
      const href = row.getAttribute("data-file-open") || "";
      if (href) {
        window.location.href = href;
      }
    });
  });

  function fileCheckboxes() {
    return Array.from(document.querySelectorAll("[data-file-select]"));
  }

  function selectedGridCards() {
    return gridCards.filter(function(card) {
      return card.classList.contains("is-selected");
    });
  }

  function selectedFileIds() {
    const ids = new Set();
    fileCheckboxes().forEach(function(checkbox) {
      if (!checkbox.checked) {
        return;
      }
      const fileId = checkbox.getAttribute("data-file-select") || "";
      if (fileId) {
        ids.add(fileId);
      }
    });
    selectedGridCards().forEach(function(card) {
      const fileId = card.getAttribute("data-file-grid-select") || "";
      if (fileId) {
        ids.add(fileId);
      }
    });
    return Array.from(ids);
  }

  function selectedFileNames(ids) {
    return ids
      .map(function(fileId) {
        const item = document.querySelector("[data-file-item='" + fileId + "']");
        return item ? String(item.getAttribute("data-file-name") || "") : "";
      })
      .filter(Boolean);
  }

  function clearGridSelection(exceptId) {
    gridCards.forEach(function(card) {
      const fileId = card.getAttribute("data-file-grid-select") || "";
      if (exceptId && fileId === exceptId) {
        return;
      }
      card.classList.remove("is-selected");
    });
  }

  function closeContextMenu() {
    activeContextFileId = "";
    if (contextMenu) {
      contextMenu.setAttribute("hidden", "hidden");
    }
  }

  function updateContextMenuMode(mode) {
    fileContextItems.forEach(function(item) {
      item.hidden = mode !== "file";
    });
    spaceContextItems.forEach(function(item) {
      item.hidden = mode !== "space";
    });
  }

  function openContextMenu(mode, fileId, x, y) {
    if (!contextMenu) {
      return;
    }
    activeContextFileId = mode === "file" ? fileId : "";
    updateContextMenuMode(mode);
    contextMenu.removeAttribute("hidden");
    const menuWidth = contextMenu.offsetWidth || 180;
    const menuHeight = contextMenu.offsetHeight || 180;
    const left = Math.min(x, window.innerWidth - menuWidth - 12);
    const top = Math.min(y, window.innerHeight - menuHeight - 12);
    contextMenu.style.left = Math.max(12, left) + "px";
    contextMenu.style.top = Math.max(12, top) + "px";
  }

  function updateSelectionState() {
    const checkboxes = fileCheckboxes();
    const checked = selectedFileIds().length;
    const allChecked = checkboxes.length > 0 && checkboxes.every(function(checkbox) { return checkbox.checked; });
    const someChecked = checked > 0;
    if (selectAll) {
      selectAll.checked = allChecked;
      selectAll.indeterminate = someChecked && !allChecked;
    }
    if (bulkDeleteButton) {
      bulkDeleteButton.disabled = checked === 0;
    }
    if (selectionCount) {
      selectionCount.textContent = checked === 1 ? "1 selected" : checked + " selected";
    }
    if (selectionToolbar) {
      if (checked > 0) {
        selectionToolbar.removeAttribute("hidden");
      } else {
        selectionToolbar.setAttribute("hidden", "hidden");
      }
    }
  }

  function removeRows(fileIds) {
    fileIds.forEach(function(fileId) {
      document.querySelectorAll("[data-file-item='" + fileId + "']").forEach(function(item) {
        if (item && item.parentNode) {
          item.parentNode.removeChild(item);
        }
      });
    });
    updateSelectionState();
  }

  function pageHasRows() {
    return document.querySelector("[data-file-item]") !== null;
  }

  function clearThumbnailCache(fileIds) {
    if (!window.ArkiveThumbnailCache || typeof window.ArkiveThumbnailCache.deleteForFiles !== "function") {
      return Promise.resolve();
    }
    return window.ArkiveThumbnailCache.deleteForFiles(fileIds);
  }

  function openDialog(fileIds, names) {
    if (!backdrop || !confirmButton) {
      return;
    }
    pendingDeleteIds = fileIds.slice();
    if (meta) {
      if (pendingDeleteIds.length === 1) {
        meta.textContent = names[0]
          ? "This will permanently delete " + names[0] + ". This action cannot be undone."
          : "This will permanently delete the file. This action cannot be undone.";
      } else {
        meta.textContent = "This will permanently delete " + pendingDeleteIds.length + " files. This action cannot be undone.";
      }
    }
    if (window.Dialog && window.Dialog.open) {
      window.Dialog.open("file-delete-backdrop");
    } else {
      backdrop.classList.remove("is-hidden");
    }
    confirmButton.disabled = false;
  }

  function closeDialog() {
    if (!backdrop) {
      return;
    }
    pendingDeleteIds = [];
    if (window.Dialog && window.Dialog.close) {
      window.Dialog.close("file-delete-backdrop");
    } else {
      backdrop.classList.add("is-hidden");
    }
  }

  deleteButtons.forEach(function(button) {
    button.addEventListener("click", function() {
      const fileId = button.getAttribute("data-file-id");
      if (!fileId) {
        return;
      }
      const filename = button.getAttribute("data-file-name") || "";
      openDialog([fileId], filename ? [filename] : []);
    });
  });

  fileCheckboxes().forEach(function(checkbox) {
    checkbox.addEventListener("change", updateSelectionState);
  });

  gridCards.forEach(function(card) {
    card.addEventListener("click", function(event) {
      if (event.target.closest("a, button, input, label")) {
        return;
      }
      const fileId = card.getAttribute("data-file-grid-select") || "";
      if (!fileId) {
        return;
      }
      if (event.ctrlKey || event.metaKey) {
        card.classList.toggle("is-selected");
      } else {
        clearGridSelection(fileId);
        card.classList.add("is-selected");
      }
      updateSelectionState();
    });

    card.addEventListener("dblclick", function() {
      const href = card.getAttribute("data-file-open") || "";
      if (href) {
        window.location.href = href;
      }
    });

    card.addEventListener("contextmenu", function(event) {
      event.preventDefault();
      const fileId = card.getAttribute("data-file-grid-select") || "";
      if (!fileId) {
        return;
      }
      if (!card.classList.contains("is-selected")) {
        clearGridSelection(fileId);
        card.classList.add("is-selected");
        updateSelectionState();
      }
      openContextMenu("file", fileId, event.clientX, event.clientY);
    });

    card.addEventListener("keydown", function(event) {
      if (event.key === "Enter") {
        event.preventDefault();
        const href = card.getAttribute("data-file-open") || "";
        if (href) {
          window.location.href = href;
        }
        return;
      }
      if (event.key !== " " && event.key !== "Spacebar") {
        return;
      }
      event.preventDefault();
      const fileId = card.getAttribute("data-file-grid-select") || "";
      if (!fileId) {
        return;
      }
      if (event.ctrlKey || event.metaKey) {
        card.classList.toggle("is-selected");
      } else {
        clearGridSelection(fileId);
        card.classList.add("is-selected");
      }
      updateSelectionState();
    });
  });

  fileContextItems.forEach(function(button) {
    button.addEventListener("click", function() {
      const action = button.getAttribute("data-grid-menu-action") || "";
      let fileName = "";
      if (activeContextFileId) {
        const item = document.querySelector("[data-file-item='" + activeContextFileId + "']");
        if (item) {
          fileName = String(item.getAttribute("data-file-name") || "");
        }
      }
      closeContextMenu();
      if (window.Toast) {
        window.Toast.success((fileName || "File") + ": " + action + " coming soon.", { title: "Context menu" });
      }
    });
  });

  spaceContextItems.forEach(function(button) {
    button.addEventListener("click", function() {
      const action = button.getAttribute("data-grid-space-action") || "";
      closeContextMenu();
      if (window.Toast) {
        window.Toast.success(action + " coming soon.", { title: "Context menu" });
      }
    });
  });

  if (gridWrap) {
    gridWrap.addEventListener("contextmenu", function(event) {
      if (event.target.closest("[data-file-grid-select]")) {
        return;
      }
      event.preventDefault();
      clearGridSelection("");
      updateSelectionState();
      openContextMenu("space", "", event.clientX, event.clientY);
    });
  }

  document.addEventListener("click", function(event) {
    if (!contextMenu || contextMenu.hasAttribute("hidden")) {
      return;
    }
    if (event.target.closest("#files-grid-context-menu")) {
      return;
    }
    closeContextMenu();
  });

  document.addEventListener("keydown", function(event) {
    if (event.key === "Escape") {
      closeContextMenu();
    }
  });

  if (selectAll) {
    selectAll.addEventListener("change", function() {
      const checked = !!selectAll.checked;
      fileCheckboxes().forEach(function(checkbox) {
        checkbox.checked = checked;
      });
      updateSelectionState();
    });
  }

  if (bulkDeleteButton) {
    bulkDeleteButton.addEventListener("click", function() {
      const ids = selectedFileIds();
      if (!ids.length) {
        return;
      }
      openDialog(ids, selectedFileNames(ids));
    });
  }

  if (cancelButton) {
    cancelButton.addEventListener("click", function() {
      closeDialog();
    });
  }

  if (confirmButton) {
    confirmButton.addEventListener("click", function() {
      if (!pendingDeleteIds.length) {
        closeDialog();
        return;
      }
      confirmButton.disabled = true;
      const isBulk = pendingDeleteIds.length > 1;
      const request = isBulk
        ? fetch("/api/files/bulk-delete", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ fileIds: pendingDeleteIds.slice() })
          })
        : fetch("/api/files/" + encodeURIComponent(pendingDeleteIds[0]), {
            method: "DELETE",
            headers: { "Content-Type": "application/json" }
          });
      request
        .then(function(res) {
          if (!res.ok) {
            throw new Error("Delete failed");
          }
          const removedIds = pendingDeleteIds.slice();
          return clearThumbnailCache(removedIds)
            .catch(function() {})
            .then(function() {
              removeRows(removedIds);
              closeDialog();
              if (!pageHasRows() || removedIds.length > 1) {
                window.location.reload();
                return;
              }
              window.Toast.success(removedIds.length === 1 ? "File deleted." : "Files deleted.", { title: "Deleted" });
            });
        })
        .catch(function() {
          window.Toast.error("Delete failed. Try again.");
          closeDialog();
        })
        .finally(function() {
          confirmButton.disabled = false;
        });
    });
  }

  updateSelectionState();
})();

(function() {
  const shareButtons = document.querySelectorAll("[data-file-action='share']");
  const backdrop = document.getElementById("file-share-backdrop");
  const linkInput = document.getElementById("share-link-input");
  const copyButton = document.getElementById("share-copy-button");
  const fileNameEl = document.getElementById("share-file-name");
  const statusEl = document.getElementById("share-status");
  const saveStateEl = document.getElementById("share-save-state");
  const saveButton = document.getElementById("share-save-button");
  const expirySelect = document.getElementById("share-expiry-select");
  const expiryToggle = document.getElementById("share-expiry-toggle");
  const expiryCustomWrap = document.querySelector(".share-expiry-custom");
  const expiryCustomInput = document.getElementById("share-expiry-custom");
  const expiryTimeInput = document.getElementById("share-expiry-time");
  const passwordToggle = document.getElementById("share-password-toggle");
  const passwordField = document.querySelector(".share-password-field");
  const passwordInput = document.getElementById("share-password");
  const passwordHelper = document.getElementById("share-password-helper");
  const passwordStrength = document.getElementById("share-password-strength");
  const deleteButton = document.getElementById("share-delete-button");
  const burnToggle = document.getElementById("share-burn-toggle");
  const closeButton = document.getElementById("share-close-button");
  const errorEl = document.getElementById("share-error");

  let activeFileId = null;
  let activeShareId = "";
  let activeShareSecret = "";

  if (!shareButtons.length || !backdrop || !saveButton || !statusEl) {
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
  }

  function setExpiryVisible(visible) {
    if (expiryCustomWrap) {
      expiryCustomWrap.classList.toggle("is-visible", !!visible);
    }
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
    const expiresAt = data && data.expiresAt ? String(data.expiresAt) : "";
    activeShareId = String((data && data.id) || "");

    if (linkInput) {
      linkInput.value = link;
    }
    if (statusEl) {
      statusEl.textContent = link ? "Live" : "Ready to create";
    }
    if (copyButton) {
      copyButton.disabled = !link;
    }
    if (deleteButton) {
      deleteButton.disabled = !activeShareId;
    }
    if (passwordToggle) {
      passwordToggle.checked = hasPassword;
    }
    if (burnToggle) {
      burnToggle.checked = false;
      burnToggle.disabled = true;
    }
    if (expiryToggle) {
      expiryToggle.checked = !!expiresAt;
    }
    setPasswordVisible(hasPassword);
    setExpiryVisible(!!expiresAt);
    updatePasswordHelper(hasPassword);
    updatePasswordStrength("");

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
    setShareError("");
    setSaveState("", "");
    updatePasswordStrength("");
  }

  function withAbsoluteURL(data) {
    if (!data || !data.token) {
      return data;
    }
    const shareSecret = String((data && data.shareSecret) || activeShareSecret || "");
    const hash = shareSecret ? "#s=" + encodeURIComponent(shareSecret) : "";
    return Object.assign({}, data, {
      url: window.location.origin + "/s/" + data.token + hash
    });
  }

  function loadFileRecord(fileId) {
    return fetch("/api/files/" + encodeURIComponent(fileId) + "/record", {
      method: "GET",
      headers: { "Content-Type": "application/json" }
    }).then(function(res) {
      return res.json().then(function(data) {
        if (!res.ok) {
          throw new Error((data && data.error) || "Failed to load file");
        }
        return data;
      });
    });
  }

  function parseJSON(res) {
    return res.text().then(function(text) {
      if (!text) {
        return null;
      }
      try {
        return JSON.parse(text);
      } catch (_) {
        return null;
      }
    });
  }

  function normalizeAPIError(data, fallback) {
    if (data && data.errors) {
      if (data.errors.password) {
        return String(data.errors.password);
      }
      if (data.errors.expiresAt) {
        return String(data.errors.expiresAt);
      }
      const keys = Object.keys(data.errors);
      if (keys.length) {
        return String(data.errors[keys[0]] || fallback || "Request failed.");
      }
    }
    if (data && data.error) {
      return String(data.error);
    }
    return fallback || "Request failed.";
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
    if (!window.ArkiveVault || !window.ArkiveVault.prepareShare) {
      return Promise.reject(new Error("Share encryption is unavailable."));
    }
    return window.ArkiveVault.waitUntilReady()
      .then(function() {
        return loadFileRecord(fileId);
      })
      .then(function(record) {
        return window.ArkiveVault.prepareShare(record, token).then(function(prepared) {
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
    if (window.Dialog && window.Dialog.open) {
      window.Dialog.open("file-share-backdrop");
    } else {
      backdrop.classList.remove("is-hidden");
    }
    fetch("/api/files/" + encodeURIComponent(fileId) + "/share", {
      method: "GET",
      headers: { "Content-Type": "application/json" }
    })
      .then(function(res) {
        return parseJSON(res).then(function(data) {
          if (res.status === 404) {
            return null;
          }
          if (!res.ok) {
            throw new Error(normalizeAPIError(data, "Failed to load share settings."));
          }
          return data;
        });
      })
      .then(function(data) {
        if (!data) {
          return;
        }
        if (data && data.encryptedShareKey && window.ArkiveVault && window.ArkiveVault.openShareKey) {
          return window.ArkiveVault.waitUntilReady()
            .then(function() {
              return window.ArkiveVault.openShareKey(data.encryptedShareKey, data.token || "");
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

  function closeShareDialog() {
    activeFileId = null;
    if (window.Dialog && window.Dialog.close) {
      window.Dialog.close("file-share-backdrop");
    } else {
      backdrop.classList.add("is-hidden");
    }
  }

  shareButtons.forEach(function(button) {
    button.addEventListener("click", function() {
      const fileId = button.getAttribute("data-file-id") || "";
      if (!fileId) {
        return;
      }
      openShareDialog(fileId, button.getAttribute("data-file-name") || "");
    });
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
        if (window.Toast) {
          window.Toast.success("Share link copied.", { title: "Copied" });
        }
      } catch (_) {
        if (window.Toast) {
          window.Toast.error("Copy failed.");
        }
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
      setShareError("");
      setSaveState("Saving...", "saving");

      const payload = {
        password: passwordToggle && passwordToggle.checked ? String((passwordInput && passwordInput.value) || "") : "",
        expiresAt: expiryToggle && expiryToggle.checked ? currentExpiryISO() : "",
        requirePassword: !!(passwordToggle && passwordToggle.checked),
      };

      if (payload.requirePassword && (!activeShareId || payload.password)) {
        const passwordMessage = passwordValidationMessage(payload.password);
        if (passwordMessage) {
          setSaveState("Save failed", "error");
          setShareError(passwordMessage);
          updatePasswordStrength(passwordMessage);
          return;
        }
      }

      const request = activeShareId
        ? fetch("/api/shares/" + encodeURIComponent(activeShareId), {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
          })
        : (function() {
            const token = createRandomShareToken();
            return createSharePayload(activeFileId, token).then(function(cryptoPayload) {
            activeShareSecret = String((cryptoPayload && cryptoPayload.shareSecret) || "");
            return fetch("/api/files/" + encodeURIComponent(activeFileId) + "/share", {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({
                token: token,
                password: payload.password,
                expiresAt: payload.expiresAt,
                encryptedShareKey: cryptoPayload.encryptedShareKey,
                encryptedFileKeyForShare: cryptoPayload.encryptedFileKeyForShare,
              }),
            });
            });
          })();

      request
        .then(function(res) {
          return parseJSON(res).then(function(data) {
            if (!res.ok) {
              throw new Error(normalizeAPIError(data, "Failed to save share settings."));
            }
            return data;
          });
        })
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
        });
    });
  }

  if (deleteButton) {
    deleteButton.addEventListener("click", function() {
      if (!activeShareId) {
        return;
      }
      setShareError("");
      setSaveState("Deleting...", "saving");
      fetch("/api/shares/" + encodeURIComponent(activeShareId), {
        method: "DELETE",
        headers: { "Content-Type": "application/json" }
      })
        .then(function(res) {
          if (!res.ok) {
            throw new Error("Failed to delete share");
          }
          applyShareState(null);
          activeShareSecret = "";
          setSaveState("Deleted", "");
        })
        .catch(function(error) {
          setSaveState("Delete failed", "error");
          setShareError((error && error.message) || "Failed to delete share.");
        });
    });
  }

  if (closeButton) {
    closeButton.addEventListener("click", function() {
      closeShareDialog();
    });
  }
})();
