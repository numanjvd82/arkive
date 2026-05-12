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
    confirmButton.textContent = "Delete";
    confirmButton.classList.add("danger");
    confirmButton.classList.remove("secondary");
    if (meta) {
      const target = fileName ? "\"" + fileName + "\"" : "this share";
      meta.textContent = "Delete " + target + " and remove its link?";
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
      const token = button.getAttribute("data-share-token") || "";
      const hasPassword = button.getAttribute("data-share-has-password") === "true";
      const encryptedKey = button.getAttribute("data-share-encrypted-key") || "";
      const baseURL = value.indexOf("http") === 0 ? value : window.location.origin + value;
      const linkPromise = (!hasPassword && encryptedKey && window.ArkiveVault && window.ArkiveVault.openShareKey)
        ? window.ArkiveVault.waitUntilReady()
            .then(function() {
              return window.ArkiveVault.openShareKey(encryptedKey, token);
            })
            .then(function(result) {
              const shareSecret = String((result && result.shareSecret) || "");
              if (!shareSecret) {
                return baseURL;
              }
              return baseURL + "#s=" + encodeURIComponent(shareSecret);
            })
        : Promise.resolve(baseURL);
      linkPromise.then(function(fullValue) {
        return writeToClipboard(fullValue);
      })
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
      const endpoint = "/api/shares/" + encodeURIComponent(shareId);
      fetch(endpoint, { method: "DELETE" })
        .then(function(res) {
          if (!res.ok) {
            throw new Error("request failed");
          }
          removeRow(shareId);
          if (window.Toast) {
            window.Toast.success("Share deleted.", { title: "Deleted" });
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
