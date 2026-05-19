import { setButtonBusy } from "./button_state.js";

function selection() {
  return window.ArkiveEntrySelection && typeof window.ArkiveEntrySelection.getSelectedEntries === "function"
    ? window.ArkiveEntrySelection.getSelectedEntries()
    : [];
}

function folderOptions() {
  const options = [{ id: "", label: "Root" }];
  const seen = new Set([""]);
  document.querySelectorAll("[data-folder-item]").forEach(function(item) {
    const id = String(item.getAttribute("data-entry-id") || item.getAttribute("data-folder-item") || "");
    if (!id || seen.has(id)) {
      return;
    }
    seen.add(id);
    options.push({
      id: id,
      label: String(item.getAttribute("data-folder-name") || item.textContent || "Encrypted folder").trim() || "Encrypted folder"
    });
  });
  return options;
}

function openMoveDialog() {
  const selected = selection();
  if (!selected.length) {
    return;
  }
  const select = document.getElementById("move-target-folder");
  const meta = document.getElementById("move-entries-meta");
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
  if (window.Dialog && window.Dialog.open) {
    window.Dialog.open("entries-move-backdrop");
  }
}

export function initMoveEntries() {
  const openButtons = [document.getElementById("move-entries-button"), document.getElementById("move-entries-selected")].filter(Boolean);
  const cancelButton = document.getElementById("move-entries-cancel");
  const confirmButton = document.getElementById("move-entries-confirm");
  const select = document.getElementById("move-target-folder");

  openButtons.forEach(function(button) {
    if (button.hasAttribute("data-move-bound")) {
      return;
    }
    button.setAttribute("data-move-bound", "true");
    button.addEventListener("click", openMoveDialog);
  });

  if (cancelButton && !cancelButton.hasAttribute("data-move-cancel-bound")) {
    cancelButton.setAttribute("data-move-cancel-bound", "true");
    cancelButton.addEventListener("click", function() {
      if (window.Dialog && window.Dialog.close) {
        window.Dialog.close("entries-move-backdrop");
      }
    });
  }

  if (confirmButton && select && !confirmButton.hasAttribute("data-move-confirm-bound")) {
    confirmButton.setAttribute("data-move-confirm-bound", "true");
    confirmButton.addEventListener("click", async function() {
      const selected = selection();
      if (!selected.length) {
        return;
      }
      try {
        setButtonBusy(confirmButton, true, { busyText: "Moving..." });
        const response = await fetch("/api/entries/move", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({
            targetFolderId: select.value || null,
            fileIds: selected.filter(function(entry) { return entry.type === "file"; }).map(function(entry) { return entry.id; }),
            folderIds: selected.filter(function(entry) { return entry.type === "folder"; }).map(function(entry) { return entry.id; })
          })
        });
        if (!response.ok) {
          throw new Error("Move failed");
        }
        window.location.reload();
      } catch (error) {
        if (window.Toast) {
          window.Toast.error((error && error.message) || "Move failed.");
        }
      } finally {
        setButtonBusy(confirmButton, false);
      }
    });
  }
}
