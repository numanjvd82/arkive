import { setButtonBusy } from "./button_state.js";
import { entrySelection } from "./file_selection.js";
import { apiRequest } from "../lib/api.js";
import { showAppError } from "../lib/toasts.js";

const CLIPBOARD_STORAGE_KEY = "arkive:entry-clipboard:v1";
const state = {
  pendingEntries: [],
};
const clipboard = {
  mode: "",
  entries: [],
  sourceFolderId: "",
};

export const moveEntries = {
  canPasteTo: canPasteTo,
  clearClipboard: clearClipboard,
  cutEntries: setClipboardCut,
  currentTargetFolderId: currentTargetFolderId,
  hasClipboard: hasClipboard,
  getClipboardEntries: clipboardEntries,
  openDialog: openMoveDialog,
  pasteInto: pasteInto,
  submitMove: submitMove,
};

function moveState() {
  return state;
}

function selection() {
  return entrySelection.getSelectedEntries();
}

function normalizeFolderID(value) {
  return String(value || "").trim();
}

function currentFolderID() {
  const current = document.querySelector("[data-current-folder-id]");
  if (!current) {
    return "";
  }
  return normalizeFolderID(current.getAttribute("data-current-folder-id"));
}

function clipboardState() {
  return clipboard;
}

function persistClipboard() {
  const state = clipboardState();
  try {
    if (!state.mode || !Array.isArray(state.entries) || !state.entries.length) {
      window.sessionStorage.removeItem(CLIPBOARD_STORAGE_KEY);
      return;
    }
    window.sessionStorage.setItem(CLIPBOARD_STORAGE_KEY, JSON.stringify({
      mode: state.mode,
      entries: state.entries,
      sourceFolderId: state.sourceFolderId || "",
    }));
  } catch (_error) {
    return;
  }
}

function restoreClipboard() {
  const state = clipboardState();
  try {
    const raw = window.sessionStorage.getItem(CLIPBOARD_STORAGE_KEY);
    if (!raw) {
      return state;
    }
    const parsed = JSON.parse(raw);
    state.mode = parsed && parsed.mode === "cut" ? "cut" : "";
    state.entries = Array.isArray(parsed && parsed.entries) ? parsed.entries.filter(Boolean) : [];
    state.sourceFolderId = normalizeFolderID(parsed && parsed.sourceFolderId);
  } catch (_error) {
    state.mode = "";
    state.entries = [];
    state.sourceFolderId = "";
  }
  return state;
}

function syncClipboardVisuals() {
  const state = clipboardState();
  const cutIDs = new Set(
    state.mode === "cut"
      ? state.entries.map(function(entry) { return String(entry && entry.id || ""); }).filter(Boolean)
      : []
  );

  document.querySelectorAll("[data-selectable-entry]").forEach(function(entry) {
    const id = String(entry.getAttribute("data-entry-id") || "");
    const isCut = cutIDs.has(id);
    entry.setAttribute("data-entry-cut", isCut ? "true" : "false");
    entry.classList.toggle("is-cut", isCut);
  });
}

function setClipboardCut(entries) {
  const normalized = Array.isArray(entries) ? entries.filter(function(entry) {
    return entry && entry.id && entry.type;
  }).map(function(entry) {
    return { id: String(entry.id), type: String(entry.type) };
  }) : [];

  const state = clipboardState();
  state.mode = normalized.length ? "cut" : "";
  state.entries = normalized;
  state.sourceFolderId = currentFolderID();
  persistClipboard();
  syncClipboardVisuals();
  syncPasteAffordances();
}

function clearClipboard() {
  const state = clipboardState();
  state.mode = "";
  state.entries = [];
  state.sourceFolderId = "";
  persistClipboard();
  syncClipboardVisuals();
  syncPasteAffordances();
}

function clipboardEntries() {
  const state = restoreClipboard();
  return state.mode === "cut" ? state.entries.slice() : [];
}

function hasClipboard() {
  return clipboardEntries().length > 0;
}

function sameEntries(left, right) {
  const a = Array.isArray(left) ? left.filter(Boolean) : [];
  const b = Array.isArray(right) ? right.filter(Boolean) : [];
  if (a.length !== b.length) {
    return false;
  }
  const leftKeys = a.map(function(entry) {
    return String(entry.type || "") + ":" + String(entry.id || "");
  }).sort();
  const rightKeys = b.map(function(entry) {
    return String(entry.type || "") + ":" + String(entry.id || "");
  }).sort();
  return leftKeys.every(function(key, index) {
    return key === rightKeys[index];
  });
}

function canPasteTo(targetFolderId) {
  const target = normalizeFolderID(targetFolderId);
  const state = restoreClipboard();
  const entries = clipboardEntries();
  if (state.mode !== "cut" || !entries.length) {
    return false;
  }
  if (target === normalizeFolderID(state.sourceFolderId)) {
    return false;
  }
  return !entries.some(function(entry) {
    return entry && entry.type === "folder" && String(entry.id) === target;
  });
}

function currentTargetFolderId() {
  return currentFolderID();
}

function syncPasteAffordances() {
  const button = document.getElementById("empty-folder-paste");
  if (!button) {
    return;
  }
  const enabled = canPasteTo(currentTargetFolderId());
  button.hidden = !enabled;
  button.disabled = !enabled;
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
  await apiRequest("/api/entries/move", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(movePayload(entries, targetFolderId))
  }, {
    code: "validation_failed",
    message: "Move failed",
  });
}

async function pasteInto(targetFolderId) {
  const entries = clipboardEntries();
  if (!entries.length) {
    showAppError(null, {
      code: "unknown_error",
      message: "Nothing to paste.",
    });
    return false;
  }
  if (!canPasteTo(targetFolderId)) {
    showAppError(null, {
      code: "unknown_error",
      message: "Nothing to move here.",
    });
    return false;
  }
  await submitMove(entries, normalizeFolderID(targetFolderId) || null);
  clearClipboard();
  if (window.Toast) {
    window.Toast.success("Moved " + entries.length + (entries.length === 1 ? " item." : " items."), {
      title: "Move complete"
    });
  }
  window.location.reload();
  return true;
}

export function initMoveEntries() {
  const openButtons = [document.getElementById("move-entries-button"), document.getElementById("move-entries-selected")].filter(Boolean);
  const cancelButton = document.getElementById("move-entries-cancel");
  const confirmButton = document.getElementById("move-entries-confirm");
  const select = document.getElementById("move-target-folder");

  restoreClipboard();
  syncClipboardVisuals();
  syncPasteAffordances();

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
        if (sameEntries(selected, clipboardEntries())) {
          clearClipboard();
        }
        window.location.reload();
      } catch (error) {
        showAppError(error, {
          code: "validation_failed",
          message: "Move failed.",
        });
      } finally {
        moveState().pendingEntries = [];
        setButtonBusy(confirmButton, false);
      }
    });
  }

  const emptyPasteButton = document.getElementById("empty-folder-paste");
  if (emptyPasteButton && !emptyPasteButton.hasAttribute("data-empty-paste-bound")) {
    emptyPasteButton.setAttribute("data-empty-paste-bound", "true");
    emptyPasteButton.addEventListener("click", async function() {
      try {
        await pasteInto(currentTargetFolderId());
      } catch (error) {
        showAppError(error, {
          code: "validation_failed",
          message: "Paste failed.",
        });
      }
    });
  }
}
