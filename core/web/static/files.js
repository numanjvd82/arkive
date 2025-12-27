(function() {
  const deleteButtons = document.querySelectorAll("[data-file-action='delete']");
  const downloadButtons = document.querySelectorAll("[data-file-action='download']");
  const backdrop = document.getElementById("file-delete-backdrop");
  const meta = document.getElementById("file-delete-meta");
  const cancelButton = document.getElementById("file-delete-cancel");
  const confirmButton = document.getElementById("file-delete-confirm");
  let pendingDelete = null;

  if (!deleteButtons.length) {
    if (!downloadButtons.length) {
      return;
    }
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

  downloadButtons.forEach(function(button) {
    button.addEventListener("click", function() {
      const fileId = button.getAttribute("data-file-id");
      if (!fileId) {
        return;
      }
      const popup = window.open("", "_blank", "noopener");
      button.disabled = true;
      fetch("/api/files/" + encodeURIComponent(fileId) + "/download", {
        method: "GET",
        headers: { "Content-Type": "application/json" }
      })
        .then(function(res) {
          if (!res.ok) {
            throw new Error("Download failed");
          }
          return res.json();
        })
        .then(function(payload) {
          if (!payload || !payload.url) {
            throw new Error("Download failed");
          }
          if (popup && !popup.closed) {
            popup.location.href = payload.url;
          } else {
            window.location.href = payload.url;
          }
        })
        .catch(function() {
          window.alert("Download failed. Try again.");
        })
        .finally(function() {
          button.disabled = false;
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
        })
        .catch(function() {
          window.alert("Delete failed. Try again.");
        })
        .finally(function() {
          confirmButton.disabled = false;
        });
    });
  }
})();
