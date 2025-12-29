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
