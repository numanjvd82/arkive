(function() {
  const rows = document.querySelectorAll("[data-file-row]");
  if (!rows.length || !window.ArkiveFileReader || !window.ArkiveVault) {
    return;
  }

  function isPreviewable(mime) {
    const value = String(mime || "").toLowerCase();
    return (
      value.startsWith("image/") ||
      value.startsWith("video/") ||
      value.startsWith("text/") ||
      value.includes("json")
    );
  }

  function updateRow(row, metadata) {
    const nameEl = row.querySelector("[data-file-field='name']");
    const typeEl = row.querySelector("[data-file-field='type']");
    const viewEl = row.querySelector("[data-file-action='view']");
    const shareEl = row.querySelector("[data-file-action='share']");
    const deleteEl = row.querySelector("[data-file-action='delete']");
    const realName = metadata && metadata.name ? metadata.name : "";
    const realType = metadata && metadata.mime ? metadata.mime : "";

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
    if (shareEl) {
      shareEl.setAttribute("data-file-name", realName);
    }
    if (deleteEl) {
      deleteEl.setAttribute("data-file-name", realName);
    }
    row.setAttribute("data-file-name", realName);
    if (viewEl) {
      if (isPreviewable(realType)) {
        viewEl.classList.remove("is-disabled");
        viewEl.removeAttribute("aria-disabled");
        viewEl.removeAttribute("title");
      } else {
        viewEl.classList.add("is-disabled");
        viewEl.setAttribute("aria-disabled", "true");
        viewEl.setAttribute("title", "Preview unavailable");
      }
    }

    row.classList.add("is-hydrated");
    row.removeAttribute("aria-busy");
  }

  async function hydrateRow(row) {
    const fileId = row.getAttribute("data-file-row");
    if (!fileId) {
      return;
    }
    const reader = new window.ArkiveFileReader({ fileId: fileId });
    try {
      await reader.load();
      updateRow(row, reader.getMetadata());
    } catch (_) {
    } finally {
      await reader.dispose();
    }
  }

  async function hydrateRows() {
    if (typeof window.ArkiveVault.waitUntilReady === "function") {
      await window.ArkiveVault.waitUntilReady();
    }
    for (let i = 0; i < rows.length; i++) {
      await hydrateRow(rows[i]);
    }
  }

  document.addEventListener("click", async function(event) {
    const target = event.target.closest("[data-file-action='download']");
    if (!target) {
      return;
    }
    event.preventDefault();
    const fileId = target.getAttribute("data-file-id");
    if (!fileId) {
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

  hydrateRows();
})();

(function() {
  const deleteButtons = document.querySelectorAll("[data-file-action='delete']");
  const selectAll = document.getElementById("files-select-all");
  const bulkDeleteButton = document.getElementById("files-delete-selected");
  const selectionCount = document.getElementById("files-selection-count");
  const backdrop = document.getElementById("file-delete-backdrop");
  const meta = document.getElementById("file-delete-meta");
  const cancelButton = document.getElementById("file-delete-cancel");
  const confirmButton = document.getElementById("file-delete-confirm");
  let pendingDeleteIds = [];

  if (!deleteButtons.length && !bulkDeleteButton) {
    return;
  }

  function fileCheckboxes() {
    return Array.from(document.querySelectorAll("[data-file-select]"));
  }

  function selectedFileIds() {
    return fileCheckboxes()
      .filter(function(checkbox) { return checkbox.checked; })
      .map(function(checkbox) { return checkbox.getAttribute("data-file-select") || ""; })
      .filter(Boolean);
  }

  function selectedFileNames(ids) {
    return ids
      .map(function(fileId) {
        const row = document.querySelector("[data-file-row='" + fileId + "']");
        return row ? String(row.getAttribute("data-file-name") || "") : "";
      })
      .filter(Boolean);
  }

  function updateSelectionState() {
    const checkboxes = fileCheckboxes();
    const checked = checkboxes.filter(function(checkbox) { return checkbox.checked; }).length;
    if (selectAll) {
      selectAll.checked = checkboxes.length > 0 && checked === checkboxes.length;
      selectAll.indeterminate = checked > 0 && checked < checkboxes.length;
    }
    if (bulkDeleteButton) {
      bulkDeleteButton.disabled = checked === 0;
    }
    if (selectionCount) {
      selectionCount.textContent = checked === 1 ? "1 selected" : checked + " selected";
    }
  }

  function removeRows(fileIds) {
    fileIds.forEach(function(fileId) {
      const row = document.querySelector("[data-file-row='" + fileId + "']");
      if (row && row.parentNode) {
        row.parentNode.removeChild(row);
      }
    });
    updateSelectionState();
  }

  function pageHasRows() {
    return document.querySelector("[data-file-row]") !== null;
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
          removeRows(removedIds);
          closeDialog();
          if (!pageHasRows() || removedIds.length > 1) {
            window.location.reload();
            return;
          }
          window.Toast.success(removedIds.length === 1 ? "File deleted." : "Files deleted.", { title: "Deleted" });
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
  const deleteButton = document.getElementById("share-delete-button");
  const burnToggle = document.getElementById("share-burn-toggle");
  const closeButton = document.getElementById("share-close-button");
  const errorEl = document.getElementById("share-error");

  let activeFileId = null;
  let activeShare = null;
  let activeFileName = "";
  let saveInFlight = false;

  if (!shareButtons.length || !backdrop) {
    return;
  }

  function openDialog() {
    if (window.Dialog && window.Dialog.open) {
      window.Dialog.open("file-share-backdrop");
    } else {
      backdrop.classList.remove("is-hidden");
    }
  }

  function closeDialog() {
    if (window.Dialog && window.Dialog.close) {
      window.Dialog.close("file-share-backdrop");
    } else {
      backdrop.classList.add("is-hidden");
    }
  }

  function setError(message) {
    if (!errorEl) {
      return;
    }
    if (!message) {
      errorEl.textContent = "";
      errorEl.classList.remove("is-visible");
      return;
    }
    errorEl.textContent = message;
    errorEl.classList.add("is-visible");
  }

  function setFileName(name) {
    activeFileName = name || "";
    if (fileNameEl) {
      fileNameEl.textContent = activeFileName || "Selected file";
    }
  }

  function setStatus(message) {
    if (statusEl) {
      statusEl.textContent = message;
    }
  }

  function setSaveState(message, stateClass) {
    if (!saveStateEl) {
      return;
    }
    saveStateEl.textContent = message || "";
    saveStateEl.classList.remove("is-saving", "is-error");
    saveStateEl.classList.remove("is-empty");
    if (stateClass) {
      saveStateEl.classList.add(stateClass);
    }
    if (!message) {
      saveStateEl.classList.add("is-empty");
    }
  }

  function setLink(token) {
    if (!linkInput) {
      return;
    }
    if (!token) {
      linkInput.value = "";
      linkInput.placeholder = "Link will appear after saving";
      return;
    }
    linkInput.value = window.location.origin + "/s/" + token;
  }

  function disableActions(isDisabled) {
    if (copyButton) {
      copyButton.disabled = isDisabled;
    }
    if (deleteButton) {
      deleteButton.disabled = isDisabled;
    }
    if (saveButton) {
      saveButton.disabled = isDisabled;
    }
    if (passwordToggle) {
      passwordToggle.disabled = isDisabled;
    }
    if (expiryToggle) {
      expiryToggle.disabled = isDisabled;
    }
    if (expirySelect) {
      expirySelect.disabled = isDisabled || !(expiryToggle && expiryToggle.checked);
    }
    if (burnToggle) {
      burnToggle.disabled = true;
    }
    if (passwordInput) {
      passwordInput.disabled = isDisabled || !(passwordToggle && passwordToggle.checked);
    }
    if (expiryCustomInput) {
      expiryCustomInput.disabled = isDisabled || !(expiryToggle && expiryToggle.checked);
    }
    if (expiryTimeInput) {
      expiryTimeInput.disabled = isDisabled || !(expiryToggle && expiryToggle.checked);
    }
  }

  function updateActionAvailability() {
    if (!activeShare) {
      disableActions(false);
      if (copyButton) {
        copyButton.disabled = true;
      }
      if (deleteButton) {
        deleteButton.disabled = true;
      }
      return;
    }
    disableActions(false);
    const revoked = activeShare.status === "revoked";
    const expired = activeShare.status === "expired" || activeShare.isExpired;
    if (copyButton) {
      copyButton.disabled = revoked || expired;
    }
    if (deleteButton) {
      deleteButton.disabled = false;
    }
    if (saveButton) {
      saveButton.disabled = false;
    }
  }

  function resetForm() {
    activeShare = null;
    setLink("");
    setStatus("Not shared");
    setSaveState("");
    setError("");
    if (expirySelect) {
      expirySelect.value = "custom";
    }
    if (expiryToggle) {
      expiryToggle.checked = false;
    }
    if (expiryCustomWrap) {
      expiryCustomWrap.classList.remove("is-visible");
    }
    if (expiryCustomInput) {
      expiryCustomInput.value = "";
    }
    if (expiryTimeInput) {
      expiryTimeInput.value = "";
    }
    if (passwordToggle) {
      passwordToggle.checked = false;
    }
    if (passwordField) {
      passwordField.classList.remove("is-visible");
    }
    if (passwordInput) {
      passwordInput.value = "";
    }
    if (passwordHelper) {
      passwordHelper.classList.remove("is-visible");
    }
    if (burnToggle) {
      burnToggle.checked = false;
      burnToggle.disabled = true;
    }
    updateActionAvailability();
  }

  function toLocalParts(isoString) {
    const date = new Date(isoString);
    if (Number.isNaN(date.getTime())) {
      return { date: "", time: "" };
    }
    const pad = function(value) {
      return value < 10 ? "0" + value : "" + value;
    };
    return {
      date: date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()),
      time: pad(date.getHours()) + ":" + pad(date.getMinutes())
    };
  }

  function applyShareState(share) {
    activeShare = share;
    setLink(share.token);
    if (share.status === "revoked") {
      setStatus("Revoked");
    } else if (share.status === "expired" || share.isExpired) {
      setStatus("Expired");
    } else {
      setStatus("Active");
    }
    setSaveState("Saved");
    if (expirySelect) {
      if (share.expiresAt) {
        expirySelect.value = "custom";
        if (expiryToggle) {
          expiryToggle.checked = true;
        }
        if (expiryCustomWrap) {
          expiryCustomWrap.classList.add("is-visible");
        }
        const parts = toLocalParts(share.expiresAt);
        if (expiryCustomInput) {
          expiryCustomInput.value = parts.date;
        }
        if (expiryTimeInput) {
          expiryTimeInput.value = parts.time;
        }
      } else {
        if (expiryToggle) {
          expiryToggle.checked = false;
        }
        if (expiryCustomWrap) {
          expiryCustomWrap.classList.remove("is-visible");
        }
      }
    }
    if (passwordToggle) {
      passwordToggle.checked = share.hasPassword || false;
    }
    if (passwordField) {
      if (share.hasPassword) {
        passwordField.classList.add("is-visible");
      } else {
        passwordField.classList.remove("is-visible");
      }
    }
    if (passwordInput) {
      passwordInput.value = "";
    }
    if (passwordHelper) {
      passwordHelper.classList.toggle("is-visible", !!share.hasPassword);
    }
    updateActionAvailability();
  }

  function buildExpiry() {
    if (!expiryToggle || !expiryToggle.checked) {
      return "";
    }
    if (!expirySelect) {
      return "";
    }
    const value = expirySelect.value;
    if (value === "custom") {
      if (!expiryCustomInput || !expiryTimeInput || !expiryCustomInput.value || !expiryTimeInput.value) {
        return "";
      }
      const customDate = new Date(expiryCustomInput.value + "T" + expiryTimeInput.value);
      if (Number.isNaN(customDate.getTime())) {
        return "";
      }
      return customDate.toISOString();
    }
    const now = Date.now();
    const offset =
      value === "1d"
        ? 24 * 60 * 60 * 1000
        : value === "7d"
        ? 7 * 24 * 60 * 60 * 1000
        : 30 * 24 * 60 * 60 * 1000;
    return new Date(now + offset).toISOString();
  }

  function fetchShare() {
    return fetch("/api/files/" + encodeURIComponent(activeFileId) + "/share", {
      method: "GET",
      headers: { "Content-Type": "application/json" }
    }).then(function(res) {
      if (!res.ok) {
        throw res;
      }
      return res.json();
    });
  }

  function createShare() {
    const passwordEnabled = passwordToggle && passwordToggle.checked;
    const password = passwordEnabled && passwordInput ? passwordInput.value : "";
    const expiresAt = buildExpiry();
    if (passwordEnabled && !password) {
      return Promise.reject(new Error("Password is required"));
    }
    if (expiryToggle && expiryToggle.checked && expirySelect && expirySelect.value === "custom" && !expiresAt) {
      return Promise.reject(new Error("Custom expiry is required"));
    }
    return fetch("/api/files/" + encodeURIComponent(activeFileId) + "/share", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        expiresAt: expiresAt,
        password: password
      })
    }).then(function(res) {
      if (!res.ok) {
        return res.json().then(function(data) {
          const error = (data && data.error) || "Share failed";
          const errors = data && data.errors;
          const message = errors && (errors.password || errors.expiresAt || errors.token);
          throw new Error(message || error);
        });
      }
      return res.json();
    });
  }

  function updateShare() {
    if (!activeShare || !activeShare.id) {
      return Promise.resolve();
    }
    const passwordEnabled = passwordToggle && passwordToggle.checked;
    const password = passwordEnabled && passwordInput ? passwordInput.value : "";
    const expiresAt = buildExpiry();
    if (passwordEnabled && !password) {
      if (!(activeShare && activeShare.hasPassword)) {
        return Promise.reject(new Error("Password is required"));
      }
    }
    if (expiryToggle && expiryToggle.checked && expirySelect && expirySelect.value === "custom" && !expiresAt) {
      return Promise.reject(new Error("Custom expiry is required"));
    }
    return fetch("/api/shares/" + encodeURIComponent(activeShare.id), {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        expiresAt: expiresAt,
        password: password,
        requirePassword: passwordEnabled
      })
    }).then(function(res) {
      if (!res.ok) {
        return res.json().then(function(data) {
          const error = (data && data.error) || "Update failed";
          const errors = data && data.errors;
          const message = errors && (errors.password || errors.expiresAt);
          throw new Error(message || error);
        });
      }
      return res.json();
    });
  }

  function deleteShare() {
    if (!activeShare || !activeShare.id) {
      return Promise.resolve();
    }
    return fetch("/api/shares/" + encodeURIComponent(activeShare.id), {
      method: "DELETE",
      headers: { "Content-Type": "application/json" }
    }).then(function(res) {
      if (!res.ok) {
        throw new Error("Delete failed");
      }
    });
  }

  function handleCreateFlow() {
    setStatus("Not shared");
    setSaveState("");
    setLink("");
    disableActions(false);
    updateActionAvailability();
  }

  shareButtons.forEach(function(button) {
    button.addEventListener("click", function() {
      activeFileId = button.getAttribute("data-file-id");
      setFileName(button.getAttribute("data-file-name") || "");
      if (!activeFileId) {
        return;
      }
      resetForm();
      setFileName(button.getAttribute("data-file-name") || "");
      disableActions(true);
      openDialog();
      fetchShare()
        .then(function(share) {
          applyShareState(share);
          disableActions(false);
        })
        .catch(function(err) {
          if (err && err.status === 404) {
            handleCreateFlow();
            return;
          }
          setStatus("Unavailable");
          setError("Unable to load share.");
          disableActions(false);
          updateActionAvailability();
        });
    });
  });

  if (passwordToggle) {
    passwordToggle.addEventListener("change", function() {
      setError("");
      if (passwordField) {
        if (passwordToggle.checked) {
          passwordField.classList.add("is-visible");
        } else {
          passwordField.classList.remove("is-visible");
          if (passwordInput) {
            passwordInput.value = "";
          }
        }
      }
      if (passwordHelper) {
        passwordHelper.classList.toggle("is-visible", !!(activeShare && activeShare.hasPassword && passwordToggle.checked));
      }
      if (passwordInput) {
        passwordInput.disabled = !(passwordToggle && passwordToggle.checked);
        if (passwordToggle.checked) {
          passwordInput.focus();
        }
      }
    });
  }

  if (expiryToggle) {
    expiryToggle.addEventListener("change", function() {
      setError("");
      if (expiryCustomWrap) {
        expiryCustomWrap.classList.toggle("is-visible", expiryToggle.checked);
      }
      if (expiryCustomInput) {
        expiryCustomInput.disabled = !expiryToggle.checked;
        if (!expiryToggle.checked) {
          expiryCustomInput.value = "";
        }
      }
      if (expiryTimeInput) {
        expiryTimeInput.disabled = !expiryToggle.checked;
        if (!expiryToggle.checked) {
          expiryTimeInput.value = "";
        }
      }
    });
  }

  if (expirySelect) {
    expirySelect.addEventListener("change", function() {
      setError("");
      if (expirySelect.value !== "custom" && expiryToggle && expiryToggle.checked) {
        if (expiryCustomInput) {
          expiryCustomInput.value = "";
        }
        if (expiryTimeInput) {
          expiryTimeInput.value = "";
        }
      }
    });
  }

  if (copyButton) {
    copyButton.addEventListener("click", function() {
      if (!linkInput || !linkInput.value) {
        return;
      }
      const value = linkInput.value;
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(value).then(function() {
          window.Toast.success("Share link copied.", { title: "Copied" });
        });
      } else {
        linkInput.select();
        document.execCommand("copy");
        window.Toast.success("Share link copied.", { title: "Copied" });
      }
    });
  }

  if (saveButton) {
    saveButton.addEventListener("click", function() {
      if (!activeFileId || saveInFlight) {
        return;
      }
      setError("");
      setSaveState("Saving...", "is-saving");
      disableActions(true);
      saveInFlight = true;
      const action = activeShare && activeShare.id ? updateShare() : createShare();
      action
        .then(function(share) {
          applyShareState(share);
          setSaveState("Saved");
          disableActions(false);
          window.Toast.success("Share settings updated.", { title: "Saved" });
        })
        .catch(function(err) {
          setSaveState("Not saved", "is-error");
          setError(err.message || "Unable to update share.");
          disableActions(false);
          updateActionAvailability();
        })
        .finally(function() {
          saveInFlight = false;
        });
    });
  }

  if (deleteButton) {
    deleteButton.addEventListener("click", function() {
      if (!activeShare || !activeShare.id) {
        return;
      }
      disableActions(true);
      setSaveState("Deleting...", "is-saving");
      deleteShare()
        .then(function() {
          resetForm();
          setStatus("Deleted");
          setSaveState("Saved");
          disableActions(false);
          updateActionAvailability();
          window.Toast.success("Share link deleted.", { title: "Deleted" });
        })
        .catch(function() {
          disableActions(false);
          setSaveState("Not saved", "is-error");
          window.Toast.error("Delete failed. Try again.");
        });
    });
  }

  if (closeButton) {
    closeButton.addEventListener("click", function() {
      closeDialog();
    });
  }
})();
