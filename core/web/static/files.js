(function() {
  const deleteButtons = document.querySelectorAll("[data-file-action='delete']");
  const backdrop = document.getElementById("file-delete-backdrop");
  const meta = document.getElementById("file-delete-meta");
  const cancelButton = document.getElementById("file-delete-cancel");
  const confirmButton = document.getElementById("file-delete-confirm");
  let pendingDelete = null;

  if (!deleteButtons.length) {
    return;
  }

  function removeRow(fileId) {
    const row = document.querySelector("[data-file-row='" + fileId + "']");
    if (row && row.parentNode) {
      row.parentNode.removeChild(row);
    }
  }

  function openDialog(fileId, filename) {
    if (!backdrop || !confirmButton) {
      return;
    }
    pendingDelete = fileId;
    if (meta) {
      meta.textContent = filename
        ? "This will permanently delete " + filename + ". This action cannot be undone."
        : "This will permanently delete the file. This action cannot be undone.";
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
    pendingDelete = null;
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
      openDialog(fileId, filename);
    });
  });

  if (cancelButton) {
    cancelButton.addEventListener("click", function() {
      closeDialog();
    });
  }

  if (confirmButton) {
    confirmButton.addEventListener("click", function() {
      if (!pendingDelete) {
        closeDialog();
        return;
      }
      confirmButton.disabled = true;
      fetch("/api/files/" + encodeURIComponent(pendingDelete), {
        method: "DELETE",
        headers: { "Content-Type": "application/json" }
      })
        .then(function(res) {
          if (!res.ok) {
            throw new Error("Delete failed");
          }
          removeRow(pendingDelete);
          closeDialog();
          window.Toast.success("File deleted.", { title: "Deleted" });
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
})();

(function() {
  const shareButtons = document.querySelectorAll("[data-file-action='share']");
  const backdrop = document.getElementById("file-share-backdrop");
  const linkInput = document.getElementById("share-link-input");
  const copyButton = document.getElementById("share-copy-button");
  const statusEl = document.getElementById("share-status");
  const expirySelect = document.getElementById("share-expiry-select");
  const expiryCustomWrap = document.querySelector(".share-expiry-custom");
  const expiryCustomInput = document.getElementById("share-expiry-custom");
  const passwordToggle = document.getElementById("share-password-toggle");
  const passwordField = document.querySelector(".share-password-field");
  const passwordInput = document.getElementById("share-password");
  const confirmButton = document.getElementById("share-confirm-button");
  const resetButton = document.getElementById("share-reset-button");
  const revokeButton = document.getElementById("share-revoke-button");
  const deleteButton = document.getElementById("share-delete-button");
  const closeButton = document.getElementById("share-close-button");
  const errorEl = document.getElementById("share-error");

  let activeFileId = null;
  let activeShare = null;
  const shareStatusEls = document.querySelectorAll("[data-file-share-status]");
  const shareExpiryEls = document.querySelectorAll("[data-file-share-expiry]");
  const shareStatusByFile = {};
  const shareExpiryByFile = {};

  if (!shareButtons.length || !backdrop) {
    return;
  }

  shareStatusEls.forEach(function(node) {
    const fileId = node.getAttribute("data-file-share-status");
    if (fileId) {
      shareStatusByFile[fileId] = node;
    }
  });

  shareExpiryEls.forEach(function(node) {
    const fileId = node.getAttribute("data-file-share-expiry");
    if (fileId) {
      shareExpiryByFile[fileId] = node;
    }
  });

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

  function setStatus(message) {
    if (statusEl) {
      statusEl.textContent = message;
    }
  }

  function setLink(token) {
    if (!linkInput) {
      return;
    }
    if (!token) {
      linkInput.value = "";
      linkInput.placeholder = "Generating link...";
      return;
    }
    linkInput.value = window.location.origin + "/s/" + token;
  }

  function formatExpiry(isoString, expired) {
    if (!isoString) {
      return "";
    }
    const date = new Date(isoString);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    const label = expired ? "Expired on " : "Expires ";
    return label + date.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric"
    });
  }

  function updateShareRow(fileId, share) {
    const statusEl = shareStatusByFile[fileId];
    const expiryEl = shareExpiryByFile[fileId];
    if (!statusEl) {
      return;
    }
    if (!share) {
      statusEl.textContent = "Not shared";
      if (expiryEl) {
        expiryEl.textContent = "";
      }
      return;
    }
    if (share.status === "revoked") {
      statusEl.textContent = "Revoked";
      if (expiryEl) {
        expiryEl.textContent = "";
      }
      return;
    }
    if (share.status === "expired" || share.isExpired) {
      statusEl.textContent = "Expired";
      if (expiryEl) {
        expiryEl.textContent = formatExpiry(share.expiresAt, true);
      }
      return;
    }
    statusEl.textContent = "Shared";
    if (expiryEl) {
      expiryEl.textContent = share.expiresAt ? formatExpiry(share.expiresAt, false) : "No expiry";
    }
  }

  function disableActions(isDisabled) {
    if (copyButton) {
      copyButton.disabled = isDisabled;
    }
    if (confirmButton) {
      confirmButton.disabled = isDisabled;
    }
    if (resetButton) {
      resetButton.disabled = isDisabled;
    }
    if (revokeButton) {
      revokeButton.disabled = isDisabled;
    }
    if (deleteButton) {
      deleteButton.disabled = isDisabled;
    }
  }

  function updateActionAvailability() {
    if (!activeShare) {
      if (copyButton) {
        copyButton.disabled = true;
      }
      if (confirmButton) {
        confirmButton.disabled = false;
      }
      if (resetButton) {
        resetButton.disabled = true;
      }
      if (revokeButton) {
        revokeButton.disabled = true;
      }
      if (deleteButton) {
        deleteButton.disabled = true;
      }
      return;
    }
    const revoked = activeShare.status === "revoked";
    const expired = activeShare.status === "expired" || activeShare.isExpired;
    if (copyButton) {
      copyButton.disabled = revoked || expired;
    }
    if (confirmButton) {
      confirmButton.disabled = revoked;
    }
    if (resetButton) {
      resetButton.disabled = false;
    }
    if (revokeButton) {
      revokeButton.disabled = revoked;
    }
    if (deleteButton) {
      deleteButton.disabled = false;
    }
  }

  function resetForm() {
    activeShare = null;
    setLink("");
    setStatus("Preparing");
    setError("");
    if (expirySelect) {
      expirySelect.value = "never";
    }
    if (expiryCustomWrap) {
      expiryCustomWrap.classList.remove("is-visible");
    }
    if (expiryCustomInput) {
      expiryCustomInput.value = "";
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
    updateActionAvailability();
  }

  function toLocalInput(isoString) {
    const date = new Date(isoString);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    const pad = function(value) {
      return value < 10 ? "0" + value : "" + value;
    };
    return (
      date.getFullYear() +
      "-" +
      pad(date.getMonth() + 1) +
      "-" +
      pad(date.getDate()) +
      "T" +
      pad(date.getHours()) +
      ":" +
      pad(date.getMinutes())
    );
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
    updateShareRow(activeFileId, share);
    if (expirySelect) {
      if (share.expiresAt) {
        expirySelect.value = "custom";
        if (expiryCustomWrap) {
          expiryCustomWrap.classList.add("is-visible");
        }
        if (expiryCustomInput) {
          expiryCustomInput.value = toLocalInput(share.expiresAt);
        }
      } else {
        expirySelect.value = "never";
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
    updateActionAvailability();
  }

  function buildExpiry() {
    if (!expirySelect) {
      return "";
    }
    const value = expirySelect.value;
    if (value === "never") {
      return "";
    }
    if (value === "custom") {
      if (!expiryCustomInput || !expiryCustomInput.value) {
        return "";
      }
      const customDate = new Date(expiryCustomInput.value);
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
    if (expirySelect && expirySelect.value === "custom" && !expiresAt) {
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
      return Promise.reject(new Error("Password is required"));
    }
    if (expirySelect && expirySelect.value === "custom" && !expiresAt) {
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

  function revokeShare() {
    if (!activeShare || !activeShare.id) {
      return Promise.resolve();
    }
    return fetch("/api/shares/" + encodeURIComponent(activeShare.id) + "/revoke", {
      method: "POST",
      headers: { "Content-Type": "application/json" }
    }).then(function(res) {
      if (!res.ok) {
        throw new Error("Revoke failed");
      }
    });
  }

  function handleCreateFlow() {
    if (!activeFileId) {
      return;
    }
    setError("");
    setStatus("Creating");
    setLink("");
    disableActions(true);
    createShare()
      .then(function(share) {
        applyShareState(share);
        disableActions(false);
      })
      .catch(function(err) {
        setStatus("Unavailable");
        setError(err.message || "Unable to create share.");
        disableActions(false);
        updateActionAvailability();
        updateShareRow(activeFileId, null);
      });
  }

  shareButtons.forEach(function(button) {
    button.addEventListener("click", function() {
      activeFileId = button.getAttribute("data-file-id");
      if (!activeFileId) {
        return;
      }
      resetForm();
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

  Object.keys(shareStatusByFile).forEach(function(fileId) {
    fetch("/api/files/" + encodeURIComponent(fileId) + "/share", {
      method: "GET",
      headers: { "Content-Type": "application/json" }
    })
      .then(function(res) {
        if (!res.ok) {
          if (res.status === 404) {
            return null;
          }
          throw new Error("Share load failed");
        }
        return res.json();
      })
      .then(function(share) {
        updateShareRow(fileId, share);
      })
      .catch(function() {
        updateShareRow(fileId, null);
      });
  });

  if (expirySelect) {
    expirySelect.addEventListener("change", function() {
      if (!expiryCustomWrap) {
        return;
      }
      if (expirySelect.value === "custom") {
        expiryCustomWrap.classList.add("is-visible");
      } else {
        expiryCustomWrap.classList.remove("is-visible");
      }
    });
  }

  if (passwordToggle) {
    passwordToggle.addEventListener("change", function() {
      if (!passwordField) {
        return;
      }
      if (passwordToggle.checked) {
        passwordField.classList.add("is-visible");
      } else {
        passwordField.classList.remove("is-visible");
        if (passwordInput) {
          passwordInput.value = "";
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

  if (confirmButton) {
    confirmButton.addEventListener("click", function() {
      if (!activeFileId) {
        return;
      }
      setError("");
      disableActions(true);
      const action = activeShare && activeShare.id ? updateShare() : createShare();
      action
        .then(function(share) {
          applyShareState(share);
          disableActions(false);
          window.Toast.success("Share updated.", { title: "Saved" });
        })
        .catch(function(err) {
          disableActions(false);
          setError(err.message || "Unable to update share.");
        });
    });
  }

  if (resetButton) {
    resetButton.addEventListener("click", function() {
      if (!activeFileId) {
        return;
      }
      setError("");
      disableActions(true);
      Promise.resolve()
        .then(function() {
          if (activeShare && activeShare.id) {
            return deleteShare();
          }
          return null;
        })
        .then(function() {
          return createShare();
        })
        .then(function(share) {
          applyShareState(share);
          disableActions(false);
          window.Toast.success("Share link reset.", { title: "Reset" });
        })
        .catch(function(err) {
          disableActions(false);
          setError(err.message || "Unable to reset share.");
        });
    });
  }

  if (revokeButton) {
    revokeButton.addEventListener("click", function() {
      if (!activeShare || !activeShare.id) {
        return;
      }
      disableActions(true);
      revokeShare()
        .then(function() {
          activeShare.status = "revoked";
          setStatus("Revoked");
          disableActions(false);
          updateActionAvailability();
          updateShareRow(activeFileId, activeShare);
          window.Toast.success("Share link revoked.", { title: "Revoked" });
        })
        .catch(function() {
          disableActions(false);
          window.Toast.error("Revoke failed. Try again.");
        });
    });
  }

  if (deleteButton) {
    deleteButton.addEventListener("click", function() {
      if (!activeShare || !activeShare.id) {
        return;
      }
      disableActions(true);
      deleteShare()
        .then(function() {
          resetForm();
          setStatus("Deleted");
          disableActions(false);
          updateShareRow(activeFileId, null);
          window.Toast.success("Share link deleted.", { title: "Deleted" });
        })
        .catch(function() {
          disableActions(false);
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
