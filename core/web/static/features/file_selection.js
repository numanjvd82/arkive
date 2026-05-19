function selectedMap() {
  if (!window.__arkiveSelectedEntries) {
    window.__arkiveSelectedEntries = new Map();
  }
  return window.__arkiveSelectedEntries;
}

function selectableEntries() {
  return Array.from(document.querySelectorAll("[data-selectable-entry]"));
}

function isGridEntry(entry) {
  return entry.classList.contains("files-card");
}

function syncEntryVisual(entry, selected) {
  entry.classList.toggle("is-selected", !!selected);
  const checkbox = entry.querySelector("[data-entry-checkbox]");
  if (checkbox) {
    checkbox.checked = !!selected;
  }
}

function updateSelectionUI() {
  const entries = selectableEntries();
  const selectedEntries = selectedMap();
  const count = selectedEntries.size;
  const toolbar = document.getElementById("entries-selection-toolbar");
  const countEl = document.getElementById("entries-selection-count");
  const moveButton = document.getElementById("move-entries-button");
  const moveSelectedButton = document.getElementById("move-entries-selected");
  const selectAll = document.getElementById("entries-select-all");
  const allSelected = entries.length > 0 && entries.every(function(entry) {
    return selectedEntries.has(String(entry.getAttribute("data-entry-id") || ""));
  });

  if (toolbar) {
    if (count > 0) {
      toolbar.removeAttribute("hidden");
    } else {
      toolbar.setAttribute("hidden", "hidden");
    }
  }
  if (countEl) {
    countEl.textContent = count === 1 ? "1 selected" : count + " selected";
  }
  if (moveButton) {
    moveButton.disabled = count === 0;
  }
  if (moveSelectedButton) {
    moveSelectedButton.disabled = count === 0;
  }
  if (selectAll) {
    selectAll.checked = allSelected;
    selectAll.indeterminate = count > 0 && !allSelected;
  }

  document.dispatchEvent(new CustomEvent("arkive:selection-change", {
    detail: { selectedEntries: Array.from(selectedEntries.values()) }
  }));
}

function setSelected(entry, selected) {
  const id = String(entry.getAttribute("data-entry-id") || "");
  const type = String(entry.getAttribute("data-entry-type") || "");
  if (!id || !type) {
    return;
  }
  const entries = selectedMap();
  if (selected) {
    entries.set(id, { id: id, type: type });
  } else {
    entries.delete(id);
  }
  syncEntryVisual(entry, selected);
}

function clearSelection() {
  const entries = selectedMap();
  entries.clear();
  selectableEntries().forEach(function(entry) {
    syncEntryVisual(entry, false);
  });
  updateSelectionUI();
}

function selectAllVisible() {
  selectableEntries().forEach(function(entry) {
    setSelected(entry, true);
  });
  updateSelectionUI();
}

function requestDeleteSelected() {
  const selectedEntries = Array.from(selectedMap().values());
  if (!selectedEntries.length) {
    return;
  }
  const hasFolders = selectedEntries.some(function(entry) {
    return entry && entry.type === "folder";
  });
  if (hasFolders) {
    if (window.Toast) {
      window.Toast.error("Folder delete is not available yet.");
    }
    return;
  }
  document.dispatchEvent(new CustomEvent("arkive:entries-delete-request", {
    detail: { selectedEntries: selectedEntries }
  }));
}

export function initFileSelection() {
  const entries = selectableEntries();
  if (!entries.length) {
    return;
  }

  window.ArkiveEntrySelection = {
    clear: clearSelection,
    getSelectedEntries: function() {
      return Array.from(selectedMap().values());
    }
  };

  entries.forEach(function(entry) {
    const checkbox = entry.querySelector("[data-entry-checkbox]");
    if (checkbox && !checkbox.hasAttribute("data-entry-checkbox-bound")) {
      checkbox.setAttribute("data-entry-checkbox-bound", "true");
      checkbox.addEventListener("change", function(event) {
        setSelected(entry, !!event.target.checked);
        updateSelectionUI();
      });
    }

    if (isGridEntry(entry) && !entry.hasAttribute("data-entry-grid-bound")) {
      entry.setAttribute("data-entry-grid-bound", "true");
      entry.addEventListener("click", function(event) {
        if (event.target.closest("a, button, input, label")) {
          return;
        }
        const id = String(entry.getAttribute("data-entry-id") || "");
        const nextSelected = !selectedMap().has(id);
        if (!event.metaKey && !event.ctrlKey) {
          clearSelection();
        }
        setSelected(entry, nextSelected);
        updateSelectionUI();
      });
    }
  });

  const selectAll = document.getElementById("entries-select-all");
  if (selectAll && !selectAll.hasAttribute("data-entries-select-all-bound")) {
    selectAll.setAttribute("data-entries-select-all-bound", "true");
    selectAll.addEventListener("change", function() {
      if (selectAll.checked) {
        selectAllVisible();
        return;
      }
      clearSelection();
    });
  }

  document.addEventListener("keydown", function(event) {
    if ((event.metaKey || event.ctrlKey) && String(event.key || "").toLowerCase() === "a") {
      event.preventDefault();
      selectAllVisible();
      return;
    }
    if (event.key === "Escape") {
      clearSelection();
      return;
    }
    if (event.key === "Delete") {
      requestDeleteSelected();
    }
  });

  updateSelectionUI();
}
