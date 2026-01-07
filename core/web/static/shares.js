(function() {
  const actionButtons = document.querySelectorAll("[data-share-action]");
  const copyButtons = document.querySelectorAll("[data-share-copy]");
  const backdrop = document.getElementById("share-action-backdrop");
  const meta = document.getElementById("share-action-meta");
  const cancelButton = document.getElementById("share-action-cancel");
  const confirmButton = document.getElementById("share-action-confirm");
  let pendingAction = null;

  if (!actionButtons.length && !copyButtons.length) {
    return;
  }

  function openDialog(action, shareId, fileName) {
    if (!backdrop || !confirmButton) {
      return;
    }
    pendingAction = { action: action, shareId: shareId };
    const verb = action === "revoke" ? "Revoke" : "Delete";
    confirmButton.textContent = verb;
    confirmButton.classList.toggle("danger", action === "delete");
    confirmButton.classList.toggle("secondary", action === "revoke");
    if (meta) {
      const target = fileName ? "\"" + fileName + "\"" : "this share";
      if (action === "revoke") {
        meta.textContent = "Revoke access for " + target + "?";
      } else {
        meta.textContent = "Delete " + target + " and remove its link?";
      }
    }
    if (window.Dialog && window.Dialog.open) {
      window.Dialog.open("share-action-backdrop");
    } else {
      backdrop.classList.remove("is-hidden");
    }
  }

  function closeDialog() {
    if (!backdrop) {
      return;
    }
    pendingAction = null;
    if (window.Dialog && window.Dialog.close) {
      window.Dialog.close("share-action-backdrop");
    } else {
      backdrop.classList.add("is-hidden");
    }
  }

  function updateRowStatus(shareId, status) {
    const statusEl = document.querySelector("[data-share-status='" + shareId + "']");
    if (!statusEl) {
      return;
    }
    statusEl.classList.remove("active", "expired", "revoked");
    statusEl.classList.add(status);
    statusEl.textContent = status.charAt(0).toUpperCase() + status.slice(1);
    const revokeButton = document.querySelector("[data-share-action='revoke'][data-share-id='" + shareId + "']");
    if (revokeButton) {
      revokeButton.parentNode.removeChild(revokeButton);
    }
  }

  function removeRow(shareId) {
    const row = document.querySelector("[data-share-row='" + shareId + "']");
    if (row && row.parentNode) {
      row.parentNode.removeChild(row);
    }
  }

  actionButtons.forEach(function(button) {
    button.addEventListener("click", function() {
      const action = button.getAttribute("data-share-action");
      const shareId = button.getAttribute("data-share-id");
      if (!action || !shareId) {
        return;
      }
      const fileName = button.getAttribute("data-share-file") || "";
      openDialog(action, shareId, fileName);
    });
  });

  function writeToClipboard(text) {
    if (!text) {
      return Promise.reject();
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text);
    }
    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.setAttribute("readonly", "readonly");
    textarea.style.position = "absolute";
    textarea.style.left = "-9999px";
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
    return Promise.resolve();
  }

  copyButtons.forEach(function(button) {
    button.addEventListener("click", function() {
      const value = button.getAttribute("data-share-copy") || "";
      const fullValue = value.indexOf("http") === 0 ? value : window.location.origin + value;
      writeToClipboard(fullValue)
        .then(function() {
          if (window.Toast) {
            window.Toast.success("Link copied.", { title: "Copied" });
          }
        })
        .catch(function() {
          if (window.Toast) {
            window.Toast.error("Copy failed. Try again.");
          }
        });
    });
  });

  if (cancelButton) {
    cancelButton.addEventListener("click", function() {
      closeDialog();
    });
  }

  if (confirmButton) {
    confirmButton.addEventListener("click", function() {
      if (!pendingAction) {
        closeDialog();
        return;
      }
      const action = pendingAction.action;
      const shareId = pendingAction.shareId;
      confirmButton.disabled = true;
      const endpoint = "/api/shares/" + encodeURIComponent(shareId) + (action === "revoke" ? "/revoke" : "");
      const method = action === "revoke" ? "POST" : "DELETE";
      fetch(endpoint, { method: method })
        .then(function(res) {
          if (!res.ok) {
            throw new Error("request failed");
          }
          if (action === "revoke") {
            updateRowStatus(shareId, "revoked");
            if (window.Toast) {
              window.Toast.success("Share revoked.", { title: "Revoked" });
            }
          } else {
            removeRow(shareId);
            if (window.Toast) {
              window.Toast.success("Share deleted.", { title: "Deleted" });
            }
          }
          closeDialog();
        })
        .catch(function() {
          if (window.Toast) {
            window.Toast.error("Action failed. Try again.");
          }
          closeDialog();
        })
        .finally(function() {
          confirmButton.disabled = false;
        });
    });
  }
})();
