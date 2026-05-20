import { setButtonBusy } from "./button_state.js";

function moveState() {
  if (!window.__arkiveMoveEntriesState) {
    window.__arkiveMoveEntriesState = {
      pendingEntries: [],
    };
  }
  return window.__arkiveMoveEntriesState;
}

function selection() {
  return window.ArkiveEntrySelection && typeof window.ArkiveEntrySelection.getSelectedEntries === "function"
    ? window.ArkiveEntrySelection.getSelectedEntries()
    : [];
}

function currentFolderID() {
  const current = document.querySelector("[data-current-folder-id]");
  if (!current) {
    return "";
  }
  return String(current.getAttribute("data-current-folder-id") || "").trim();
}

function folderOptions() {
  const options = [{ id: "", label: "Root" }];
  const seen = new Set([""]);
  const currentFolderId = currentFolderID();

  function pushOption(id, label) {
    const value = String(id || "").trim();
    if (!value || seen.has(value) || value === currentFolderId) {
      return;
    }
    seen.add(value);
    options.push({
      id: value,
      label: String(label || "Encrypted folder").trim() || "Encrypted folder"
    });
  }

  document.querySelectorAll("[data-folder-breadcrumb]").forEach(function(item) {
    pushOption(
      item.getAttribute("data-folder-breadcrumb") || "",
      item.getAttribute("data-folder-name") || item.textContent || "Encrypted folder",
    );
  });

  document.querySelectorAll("[data-folder-item]").forEach(function(item) {
    pushOption(
      item.getAttribute("data-entry-id") || item.getAttribute("data-folder-item") || "",
      item.getAttribute("data-folder-name") || item.textContent || "Encrypted folder",
    );
  });
  return options;
}

function movePayload(selected, targetFolderId) {
  return {
    targetFolderId: targetFolderId || null,
    fileIds: selected.filter(function(entry) { return entry.type === "file"; }).map(function(entry) { return entry.id; }),
    folderIds: selected.filter(function(entry) { return entry.type === "folder"; }).map(function(entry) { return entry.id; })
  };
}

function openMoveDialog(entries) {
  const selected = Array.isArray(entries) && entries.length ? entries : selection();
  if (!selected.length) {
    return;
  }
  moveState().pendingEntries = selected.slice();
  const select = document.getElementById("move-target-folder");
  const meta = document.getElementById("move-entries-meta");
  const confirmButton = document.getElementById("move-entries-confirm");
  if (!select) {
    return;
  }
  select.innerHTML = "";
  folderOptions().forEach(function(option) {
    const node = document.createElement("option");
    node.value = option.id;
    node.textContent = option.label;
    select.appendChild(node);
  });
  if (meta) {
    meta.textContent = selected.length === 1 ? "Move 1 selected entry." : "Move " + selected.length + " selected entries.";
  }
  if (confirmButton) {
    setButtonBusy(confirmButton, false);
  }
  if (window.Dialog && window.Dialog.open) {
    window.Dialog.open("entries-move-backdrop");
  }
}

async function submitMove(selected, targetFolderId) {
  const entries = Array.isArray(selected) ? selected.filter(Boolean) : [];
  if (!entries.length) {
    return;
  }
  const response = await fetch("/api/entries/move", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(movePayload(entries, targetFolderId))
  });
  if (!response.ok) {
    throw new Error("Move failed");
  }
}

export function initMoveEntries() {
  const openButtons = [document.getElementById("move-entries-button"), document.getElementById("move-entries-selected")].filter(Boolean);
  const cancelButton = document.getElementById("move-entries-cancel");
  const confirmButton = document.getElementById("move-entries-confirm");
  const select = document.getElementById("move-target-folder");

  window.ArkiveMoveEntries = {
    openDialog: openMoveDialog,
    submitMove: submitMove,
  };

  openButtons.forEach(function(button) {
    if (button.hasAttribute("data-move-bound")) {
      return;
    }
    button.setAttribute("data-move-bound", "true");
    button.addEventListener("click", function() {
      openMoveDialog();
    });
  });

  if (cancelButton && !cancelButton.hasAttribute("data-move-cancel-bound")) {
    cancelButton.setAttribute("data-move-cancel-bound", "true");
    cancelButton.addEventListener("click", function() {
      moveState().pendingEntries = [];
      if (window.Dialog && window.Dialog.close) {
        window.Dialog.close("entries-move-backdrop");
      }
    });
  }

  if (confirmButton && select && !confirmButton.hasAttribute("data-move-confirm-bound")) {
    confirmButton.setAttribute("data-move-confirm-bound", "true");
    confirmButton.addEventListener("click", async function() {
      const selected = moveState().pendingEntries.length ? moveState().pendingEntries.slice() : selection();
      if (!selected.length) {
        return;
      }
      try {
        setButtonBusy(confirmButton, true, { busyText: "Moving..." });
        await submitMove(selected, select.value || null);
        window.location.reload();
      } catch (error) {
        if (window.Toast) {
          window.Toast.error((error && error.message) || "Move failed.");
        }
      } finally {
        moveState().pendingEntries = [];
        setButtonBusy(confirmButton, false);
      }
    });
  }
}
