function selectedEntries() {
  return window.ArkiveEntrySelection && typeof window.ArkiveEntrySelection.getSelectedEntries === "function"
    ? window.ArkiveEntrySelection.getSelectedEntries()
    : [];
}

function clearDragging() {
  document.querySelectorAll(".is-dragging").forEach(function(node) {
    node.classList.remove("is-dragging");
  });
  document.querySelectorAll(".is-drop-target").forEach(function(node) {
    node.classList.remove("is-drop-target");
  });
}

function targetFolderID(folder) {
  return String(folder.getAttribute("data-entry-id") || folder.getAttribute("data-folder-item") || "");
}

function invalidDrop(selection, targetID) {
  if (!targetID) {
    return false;
  }
  return selection.some(function(entry) {
    return entry.type === "folder" && entry.id === targetID;
  });
}

function markSuccess(node) {
  node.classList.add("drop-success");
  window.setTimeout(function() {
    node.classList.remove("drop-success");
  }, 320);
}

export function initEntryDragMove() {
  const entries = Array.from(document.querySelectorAll("[data-selectable-entry]"));
  const folders = Array.from(document.querySelectorAll("[data-folder-item]"));
  if (!entries.length || !folders.length) {
    return;
  }

  entries.forEach(function(entry) {
    if (entry.hasAttribute("data-entry-drag-bound")) {
      return;
    }
    entry.setAttribute("data-entry-drag-bound", "true");
    entry.setAttribute("draggable", "true");

    entry.addEventListener("dragstart", function(event) {
      const id = String(entry.getAttribute("data-entry-id") || "");
      let selection = selectedEntries();
      const alreadySelected = selection.some(function(item) {
        return item && item.id === id;
      });
      if (!alreadySelected && window.ArkiveEntrySelection && typeof window.ArkiveEntrySelection.selectOnly === "function") {
        window.ArkiveEntrySelection.selectOnly(entry);
        selection = selectedEntries();
      }
      selection.forEach(function(item) {
        const node = window.ArkiveEntrySelection && typeof window.ArkiveEntrySelection.findEntryByID === "function"
          ? window.ArkiveEntrySelection.findEntryByID(item.id)
          : null;
        if (node) {
          node.classList.add("is-dragging");
        }
      });
      event.dataTransfer.effectAllowed = "move";
      event.dataTransfer.setData("text/plain", JSON.stringify(selection));
    });

    entry.addEventListener("dragend", clearDragging);
  });

  folders.forEach(function(folder) {
    if (folder.hasAttribute("data-folder-drop-bound")) {
      return;
    }
    folder.setAttribute("data-folder-drop-bound", "true");

    folder.addEventListener("dragover", function(event) {
      const selection = selectedEntries();
      const targetID = targetFolderID(folder);
      if (!selection.length || invalidDrop(selection, targetID)) {
        return;
      }
      event.preventDefault();
      event.dataTransfer.dropEffect = "move";
      folder.classList.add("is-drop-target");
    });

    folder.addEventListener("dragleave", function() {
      folder.classList.remove("is-drop-target");
    });

    folder.addEventListener("drop", async function(event) {
      event.preventDefault();
      const selection = selectedEntries();
      const targetID = targetFolderID(folder);
      folder.classList.remove("is-drop-target");
      if (!selection.length || !targetID || invalidDrop(selection, targetID)) {
        clearDragging();
        return;
      }
      try {
        if (!window.ArkiveMoveEntries || typeof window.ArkiveMoveEntries.submitMove !== "function") {
          throw new Error("Move unavailable");
        }
        await window.ArkiveMoveEntries.submitMove(selection, targetID);
        markSuccess(folder);
        window.location.reload();
      } catch (error) {
        window.ArkiveUI.showAppError(error, {
          code: "validation_failed",
          message: "Move failed.",
        });
      } finally {
        clearDragging();
      }
    });
  });
}
