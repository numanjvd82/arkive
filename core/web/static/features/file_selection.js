import { showAppError } from "../lib/toasts.js";

import { filesActions } from "../files.js";
import { moveEntries } from "./move_entries.js";

const selectedEntriesMap = new Map();
const state = {
  lastIndex: -1,
  touchTimer: null,
  touchSelectionTriggered: false,
};

export const entrySelection = {
  clear: clearSelection,
  focusEntry: focusEntry,
  getFocusedEntry: focusedEntry,
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
  requestRenameSelected: requestRenameSelected,
  requestShareSelected: requestShareSelected,
};

function selectedMap() {
  return selectedEntriesMap;
}

function selectionState() {
  return state;
}

function selectableEntries() {
  return Array.from(document.querySelectorAll("[data-selectable-entry]"));
}

function visibleEntries() {
  return selectableEntries().filter(function(entry) {
    return entry && entry.offsetParent !== null;
  });
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

function focusedEntry() {
  const active = document.activeElement && document.activeElement.closest
    ? document.activeElement.closest("[data-selectable-entry]")
    : null;
  if (active) {
    return active;
  }
  return document.querySelector("[data-selectable-entry].is-focused") || null;
}

function updateFocusVisuals(targetEntry) {
  selectableEntries().forEach(function(entry) {
    const focused = !!targetEntry && entry === targetEntry;
    entry.classList.toggle("is-focused", focused);
    entry.tabIndex = focused ? 0 : -1;
  });
}

function focusEntry(entry) {
  if (!entry) {
    return;
  }
  updateFocusVisuals(entry);
  try {
    entry.focus({ preventScroll: true });
  } catch (_error) {
    entry.focus();
  }
}

function focusEntryAt(index) {
  const entries = visibleEntries();
  if (!entries.length) {
    return;
  }
  const clamped = Math.max(0, Math.min(index, entries.length - 1));
  focusEntry(entries[clamped]);
}

function gridColumnCount(entries) {
  if (!entries.length) {
    return 1;
  }
  const firstTop = entries[0].getBoundingClientRect().top;
  let count = 0;
  for (let index = 0; index < entries.length; index += 1) {
    if (Math.abs(entries[index].getBoundingClientRect().top - firstTop) > 12) {
      break;
    }
    count += 1;
  }
  return Math.max(1, count);
}

function moveFocusByKey(key) {
  const entries = visibleEntries();
  if (!entries.length) {
    return;
  }
  const current = focusedEntry() || entries[0];
  const index = entries.findIndex(function(entry) {
    return entry === current;
  });
  const currentIndex = index >= 0 ? index : 0;
  const inGrid = document.querySelector(".files-grid") !== null;
  const columns = inGrid ? gridColumnCount(entries) : 1;

  switch (key) {
    case "ArrowLeft":
      focusEntryAt(currentIndex - 1);
      return true;
    case "ArrowRight":
      focusEntryAt(currentIndex + 1);
      return true;
    case "ArrowUp":
      focusEntryAt(currentIndex - columns);
      return true;
    case "ArrowDown":
      focusEntryAt(currentIndex + columns);
      return true;
    case "Home":
      focusEntryAt(0);
      return true;
    case "End":
      focusEntryAt(entries.length - 1);
      return true;
    default:
      return false;
  }
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
  const entries = visibleEntries();
  const selectedEntries = selectedMap();
  const count = selectedEntries.size;
  const selectAll = document.getElementById("entries-select-all");
  const allSelected = entries.length > 0 && entries.every(function(entry) {
    return selectedEntries.has(entryID(entry));
  });
  const selectedValues = Array.from(selectedEntries.values());
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
  visibleEntries().forEach(function(entry) {
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
    showAppError(null, {
      code: "unknown_error",
      message: "Only files can be shared in this phase.",
    });
    return;
  }
  if (selectedFiles.length > 1) {
    showAppError(null, {
      code: "unknown_error",
      message: "Share currently opens one file at a time.",
    });
    return;
  }
  if (typeof filesActions.openShare === "function") {
    const node = findEntryByID(selectedFiles[0].id);
    filesActions.openShare(selectedFiles[0].id, node ? String(node.getAttribute("data-file-name") || "") : "");
  }
}

function requestMoveSelected() {
  const selected = selectedEntries();
  if (!selected.length) {
    return;
  }
  if (typeof moveEntries.openDialog === "function") {
    moveEntries.openDialog(selected);
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

function requestRenameSelected() {
  const selected = selectedEntries();
  if (selected.length > 1) {
    showAppError(null, {
      code: "unknown_error",
      message: "Rename only supports one item at a time.",
    });
    return;
  }
  document.dispatchEvent(new CustomEvent("arkive:rename-request", {
    detail: {
      entry: selected.length === 1 ? findEntryByID(selected[0].id) : focusedEntry()
    }
  }));
}

function requestCutSelected() {
  const selected = selectedEntries();
  if (!selected.length) {
    return;
  }
  moveEntries.cutEntries(selected);
  if (window.Toast) {
    window.Toast.success("Cut " + selected.length + (selected.length === 1 ? " item." : " items."), {
      title: "Ready to move"
    });
  }
}

async function requestPasteHere() {
  try {
    await moveEntries.pasteInto(
      document.querySelector("[data-current-folder-id]")
        ? document.querySelector("[data-current-folder-id]").getAttribute("data-current-folder-id") || ""
        : ""
    );
  } catch (error) {
    showAppError(error, {
      code: "validation_failed",
      message: "Paste failed.",
    });
  }
}

function clearCutClipboard() {
  if (!moveEntries.hasClipboard()) {
    return false;
  }
  moveEntries.clearClipboard();
  if (window.Toast) {
    window.Toast.success("Cut cancelled.", { title: "Clipboard cleared" });
  }
  return true;
}

function openSelectedOrFocused() {
  const selected = selectedEntries();
  const focusNode = focusedEntry();

  if (selected.length === 1) {
    const node = findEntryByID(selected[0].id);
    if (typeof filesActions.openEntry === "function" && node) {
      filesActions.openEntry(node);
    }
    return;
  }

  if (focusNode && typeof filesActions.openEntry === "function") {
    filesActions.openEntry(focusNode);
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
    focusEntry(entry);
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
    focusEntry(entry);
    if (typeof filesActions.openEntry === "function") {
      filesActions.openEntry(entry);
    }
  });

  entry.addEventListener("touchmove", function() {
    window.clearTimeout(selectionState().touchTimer);
  }, { passive: true });

  if (!entry.hasAttribute("tabindex")) {
    entry.tabIndex = -1;
  }

  entry.addEventListener("focus", function() {
    updateFocusVisuals(entry);
  });
}

function bindEmptySpaceClear() {
  if (document.body.hasAttribute("data-entry-empty-clear-bound")) {
    return;
  }
  document.body.setAttribute("data-entry-empty-clear-bound", "true");
  document.addEventListener("click", function(event) {
    if (event.target.closest("[data-selectable-entry], .files-context-menu, .dialog-backdrop, .dialog, [data-entry-checkbox], #entries-select-all, thead, th, label, button, input")) {
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
    if ((event.metaKey || event.ctrlKey) && lower === "x") {
      if (!selectedEntries().length) {
        return;
      }
      event.preventDefault();
      requestCutSelected();
      return;
    }
    if ((event.metaKey || event.ctrlKey) && lower === "v") {
      const targetFolderId = moveEntries.currentTargetFolderId();
      if (!moveEntries.canPasteTo(targetFolderId)) {
        return;
      }
      event.preventDefault();
      requestPasteHere();
      return;
    }
    if (key === "Escape") {
      if (interactionOverlayOpen()) {
        document.dispatchEvent(new CustomEvent("arkive:context-menu-close"));
        return;
      }
      if (selectedEntries().length) {
        clearSelection();
        return;
      }
      clearCutClipboard();
      return;
    }
    if (moveFocusByKey(key)) {
      event.preventDefault();
      return;
    }
    if (key === "Delete") {
      requestDeleteSelected();
      return;
    }
    if (key === "F2") {
      event.preventDefault();
      requestRenameSelected();
      return;
    }
    if (key === " " || key === "Spacebar") {
      const entry = focusedEntry();
      if (!entry) {
        return;
      }
      event.preventDefault();
      toggleEntry(entry);
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

export function initFileSelection() {
  const entries = selectableEntries();
  if (!entries.length) {
    return;
  }

  entries.forEach(bindEntry);
  bindEmptySpaceClear();
  bindShortcuts();

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

  updateSelectionUI();
}
