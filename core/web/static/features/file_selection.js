function selectedMap() {
  if (!window.__arkiveSelectedEntries) {
    window.__arkiveSelectedEntries = new Map();
  }
  return window.__arkiveSelectedEntries;
}

function selectionState() {
  if (!window.__arkiveSelectionState) {
    window.__arkiveSelectionState = {
      lastIndex: -1,
      touchTimer: null,
      touchSelectionTriggered: false,
    };
  }
  return window.__arkiveSelectionState;
}

function selectableEntries() {
  return Array.from(document.querySelectorAll("[data-selectable-entry]"));
}

function entryID(entry) {
  return String(entry && entry.getAttribute("data-entry-id") || "");
}

function entryType(entry) {
  return String(entry && entry.getAttribute("data-entry-type") || "");
}

function entryInfo(entry) {
  const id = entryID(entry);
  const type = entryType(entry);
  if (!id || !type) {
    return null;
  }
  return { id: id, type: type };
}

function entryIndex(entry) {
  return selectableEntries().findIndex(function(node) {
    return node === entry;
  });
}

function findEntryByID(id) {
  return selectableEntries().find(function(entry) {
    return entryID(entry) === String(id || "");
  }) || null;
}

function isTypingTarget(target) {
  if (!target) {
    return false;
  }
  const el = target instanceof Element ? target : null;
  if (!el) {
    return false;
  }
  if (el.closest("[contenteditable='true']")) {
    return true;
  }
  const field = el.closest("input, textarea, select");
  return !!field;
}

function isInteractiveTarget(target) {
  return !!(target && target.closest && target.closest("a, button, input, label, select, textarea"));
}

function interactionOverlayOpen() {
  return document.querySelector(".dialog-backdrop:not(.is-hidden), .files-context-menu:not([hidden])") !== null;
}

function syncEntryVisual(entry, selected) {
  entry.classList.toggle("is-selected", !!selected);
  const checkbox = entry.querySelector("[data-entry-checkbox]");
  if (checkbox) {
    checkbox.checked = !!selected;
  }
  entry.setAttribute("aria-selected", selected ? "true" : "false");
}

function updateSelectionUI() {
  const entries = selectableEntries();
  const selectedEntries = selectedMap();
  const count = selectedEntries.size;
  const toolbar = document.getElementById("entries-selection-toolbar");
  const countEl = document.getElementById("entries-selection-count");
  const moveButton = document.getElementById("move-entries-selected");
  const shareButton = document.getElementById("share-entries-selected");
  const deleteButton = document.getElementById("delete-entries-selected");
  const clearButton = document.getElementById("clear-entries-selection");
  const selectAll = document.getElementById("entries-select-all");
  const allSelected = entries.length > 0 && entries.every(function(entry) {
    return selectedEntries.has(entryID(entry));
  });
  const selectedValues = Array.from(selectedEntries.values());
  const hasFiles = selectedValues.some(function(entry) { return entry.type === "file"; });

  if (toolbar) {
    toolbar.toggleAttribute("hidden", count === 0);
  }
  if (countEl) {
    countEl.textContent = count === 1 ? "1 selected" : count + " selected";
  }
  if (moveButton) {
    moveButton.disabled = count === 0;
  }
  if (shareButton) {
    shareButton.disabled = count === 0 || !hasFiles;
  }
  if (deleteButton) {
    deleteButton.disabled = count === 0;
  }
  if (clearButton) {
    clearButton.disabled = count === 0;
  }
  if (selectAll) {
    selectAll.checked = allSelected;
    selectAll.indeterminate = count > 0 && !allSelected;
  }

  document.dispatchEvent(new CustomEvent("arkive:selection-change", {
    detail: { selectedEntries: selectedValues }
  }));
}

function setSelected(entry, selected, options) {
  const info = entryInfo(entry);
  if (!info) {
    return;
  }
  const entries = selectedMap();
  if (selected) {
    entries.set(info.id, info);
  } else {
    entries.delete(info.id);
  }
  syncEntryVisual(entry, selected);
  if (!options || !options.silent) {
    updateSelectionUI();
  }
}

function clearSelection(options) {
  selectedMap().clear();
  selectableEntries().forEach(function(entry) {
    syncEntryVisual(entry, false);
  });
  if (!options || !options.silent) {
    updateSelectionUI();
  }
}

function selectAllVisible() {
  selectableEntries().forEach(function(entry) {
    setSelected(entry, true, { silent: true });
  });
  updateSelectionUI();
}

function selectOnly(entry) {
  clearSelection({ silent: true });
  setSelected(entry, true, { silent: true });
  selectionState().lastIndex = entryIndex(entry);
  updateSelectionUI();
}

function toggleEntry(entry) {
  const id = entryID(entry);
  const next = !selectedMap().has(id);
  setSelected(entry, next, { silent: true });
  selectionState().lastIndex = entryIndex(entry);
  updateSelectionUI();
}

function selectRange(entry) {
  const entries = selectableEntries();
  const state = selectionState();
  const targetIndex = entryIndex(entry);
  const anchorIndex = state.lastIndex >= 0 ? state.lastIndex : targetIndex;
  const start = Math.min(anchorIndex, targetIndex);
  const end = Math.max(anchorIndex, targetIndex);
  clearSelection({ silent: true });
  for (let index = start; index <= end; index += 1) {
    setSelected(entries[index], true, { silent: true });
  }
  updateSelectionUI();
}

function selectedEntries() {
  return Array.from(selectedMap().values());
}

function requestDeleteSelected() {
  const selected = selectedEntries();
  if (!selected.length) {
    return;
  }
  document.dispatchEvent(new CustomEvent("arkive:entries-delete-request", {
    detail: { selectedEntries: selected }
  }));
}

function requestShareSelected() {
  const selected = selectedEntries();
  const selectedFiles = selected.filter(function(entry) { return entry.type === "file"; });
  if (!selectedFiles.length) {
    if (window.Toast) {
      window.Toast.error("Only files can be shared in this phase.");
    }
    return;
  }
  if (selectedFiles.length > 1) {
    if (window.Toast) {
      window.Toast.error("Share currently opens one file at a time.");
    }
    return;
  }
  if (window.ArkiveFilesActions && typeof window.ArkiveFilesActions.openShare === "function") {
    const node = findEntryByID(selectedFiles[0].id);
    window.ArkiveFilesActions.openShare(selectedFiles[0].id, node ? String(node.getAttribute("data-file-name") || "") : "");
  }
}

function requestMoveSelected() {
  const selected = selectedEntries();
  if (!selected.length) {
    return;
  }
  if (window.ArkiveMoveEntries && typeof window.ArkiveMoveEntries.openDialog === "function") {
    window.ArkiveMoveEntries.openDialog(selected);
    return;
  }
  const button = document.getElementById("move-entries-selected");
  if (button) {
    button.click();
  }
}

function requestNewFolder() {
  const button = document.getElementById("new-folder-button");
  if (button) {
    button.click();
  }
}

function openSelectedOrFocused() {
  const selected = selectedEntries();
  const focusEntry = document.activeElement && document.activeElement.closest
    ? document.activeElement.closest("[data-selectable-entry]")
    : null;

  if (selected.length === 1) {
    const node = findEntryByID(selected[0].id);
    if (window.ArkiveFilesActions && typeof window.ArkiveFilesActions.openEntry === "function" && node) {
      window.ArkiveFilesActions.openEntry(node);
    }
    return;
  }

  if (focusEntry && window.ArkiveFilesActions && typeof window.ArkiveFilesActions.openEntry === "function") {
    window.ArkiveFilesActions.openEntry(focusEntry);
  }
}

function applyClickSelection(entry, event) {
  if (event.shiftKey) {
    selectRange(entry);
    return;
  }
  if (event.metaKey || event.ctrlKey) {
    toggleEntry(entry);
    return;
  }
  selectOnly(entry);
}

function bindEntry(entry) {
  if (entry.hasAttribute("data-entry-selection-bound")) {
    return;
  }
  entry.setAttribute("data-entry-selection-bound", "true");

  const checkbox = entry.querySelector("[data-entry-checkbox]");
  if (checkbox && !checkbox.hasAttribute("data-entry-checkbox-bound")) {
    checkbox.setAttribute("data-entry-checkbox-bound", "true");
    checkbox.addEventListener("change", function(event) {
      setSelected(entry, !!event.target.checked);
      selectionState().lastIndex = entryIndex(entry);
    });
  }

  entry.addEventListener("click", function(event) {
    if (isInteractiveTarget(event.target)) {
      return;
    }
    applyClickSelection(entry, event);
  });

  entry.addEventListener("touchstart", function(event) {
    if (isInteractiveTarget(event.target)) {
      return;
    }
    const state = selectionState();
    state.touchSelectionTriggered = false;
    window.clearTimeout(state.touchTimer);
    state.touchTimer = window.setTimeout(function() {
      state.touchSelectionTriggered = true;
      selectOnly(entry);
    }, 420);
  }, { passive: true });

  entry.addEventListener("touchend", function(event) {
    if (isInteractiveTarget(event.target)) {
      return;
    }
    const state = selectionState();
    window.clearTimeout(state.touchTimer);
    if (state.touchSelectionTriggered) {
      event.preventDefault();
      return;
    }
    if (window.ArkiveFilesActions && typeof window.ArkiveFilesActions.openEntry === "function") {
      window.ArkiveFilesActions.openEntry(entry);
    }
  });

  entry.addEventListener("touchmove", function() {
    window.clearTimeout(selectionState().touchTimer);
  }, { passive: true });
}

function bindEmptySpaceClear() {
  if (document.body.hasAttribute("data-entry-empty-clear-bound")) {
    return;
  }
  document.body.setAttribute("data-entry-empty-clear-bound", "true");
  document.addEventListener("click", function(event) {
    if (event.target.closest("[data-selectable-entry], .files-context-menu, .dialog-backdrop, .dialog")) {
      return;
    }
    clearSelection();
  });
}

function bindShortcuts() {
  if (document.body.hasAttribute("data-entry-shortcuts-bound")) {
    return;
  }
  document.body.setAttribute("data-entry-shortcuts-bound", "true");
  document.addEventListener("keydown", function(event) {
    if (isTypingTarget(event.target)) {
      return;
    }
    const key = String(event.key || "");
    const lower = key.toLowerCase();
    if (interactionOverlayOpen()) {
      if (key !== "Escape") {
        return;
      }
    }

    if ((event.metaKey || event.ctrlKey) && lower === "a") {
      event.preventDefault();
      selectAllVisible();
      return;
    }
    if (key === "Escape") {
      clearSelection();
      document.dispatchEvent(new CustomEvent("arkive:context-menu-close"));
      return;
    }
    if (key === "Delete") {
      requestDeleteSelected();
      return;
    }
    if (key === "Enter") {
      event.preventDefault();
      openSelectedOrFocused();
      return;
    }
    if (!event.metaKey && !event.ctrlKey && !event.altKey && lower === "m") {
      if (!selectedEntries().length) {
        return;
      }
      event.preventDefault();
      requestMoveSelected();
      return;
    }
    if (!event.metaKey && !event.ctrlKey && !event.altKey && lower === "n") {
      if (!document.querySelector(".files-page") || !document.getElementById("new-folder-button")) {
        return;
      }
      event.preventDefault();
      requestNewFolder();
    }
  });
}

function bindToolbarButtons() {
  const shareButton = document.getElementById("share-entries-selected");
  const deleteButton = document.getElementById("delete-entries-selected");
  const clearButton = document.getElementById("clear-entries-selection");

  if (shareButton && !shareButton.hasAttribute("data-selection-share-bound")) {
    shareButton.setAttribute("data-selection-share-bound", "true");
    shareButton.addEventListener("click", requestShareSelected);
  }
  if (deleteButton && !deleteButton.hasAttribute("data-selection-delete-bound")) {
    deleteButton.setAttribute("data-selection-delete-bound", "true");
    deleteButton.addEventListener("click", requestDeleteSelected);
  }
  if (clearButton && !clearButton.hasAttribute("data-selection-clear-bound")) {
    clearButton.setAttribute("data-selection-clear-bound", "true");
    clearButton.addEventListener("click", function() {
      clearSelection();
    });
  }
}

export function initFileSelection() {
  const entries = selectableEntries();
  if (!entries.length) {
    return;
  }

  entries.forEach(bindEntry);
  bindEmptySpaceClear();
  bindShortcuts();
  bindToolbarButtons();

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

  window.ArkiveEntrySelection = {
    clear: clearSelection,
    getSelectedEntries: selectedEntries,
    findEntryByID: findEntryByID,
    selectOnly: selectOnly,
    setSelected: setSelected,
    replaceSelection: function(entriesToSelect) {
      clearSelection({ silent: true });
      entriesToSelect.forEach(function(entry) {
        setSelected(entry, true, { silent: true });
      });
      updateSelectionUI();
    },
    requestDeleteSelected: requestDeleteSelected,
    requestMoveSelected: requestMoveSelected,
    requestShareSelected: requestShareSelected,
  };

  updateSelectionUI();
}
