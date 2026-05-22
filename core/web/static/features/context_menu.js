import { showAppError } from "../lib/toasts.js";
import { filesActions } from "../files.js";
import { entrySelection } from "./file_selection.js";
import { moveEntries } from "./move_entries.js";

const state = {
  root: null,
  currentEntry: null,
  currentSelection: [],
};

function menuState() {
  return state;
}

function selectionAPI() {
  return entrySelection;
}

function selectedEntries() {
  const api = selectionAPI();
  return api && typeof api.getSelectedEntries === "function" ? api.getSelectedEntries() : [];
}

function ensureMenu() {
  const state = menuState();
  if (state.root) {
    return state.root;
  }
  const root = document.createElement("div");
  root.className = "files-context-menu";
  root.hidden = true;
  document.body.appendChild(root);
  state.root = root;
  return root;
}

function closeMenu() {
  const state = menuState();
  if (!state.root) {
    return;
  }
  state.root.hidden = true;
  state.root.innerHTML = "";
  state.currentEntry = null;
  state.currentSelection = [];
}

function menuButton(item, action) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = "files-context-menu-item";
  button.textContent = item.label;
  if (item.disabled) {
    button.disabled = true;
    button.setAttribute("aria-disabled", "true");
  } else {
    button.addEventListener("click", function() {
      action(item.action);
      closeMenu();
    });
  }
  return button;
}

function menuDivider() {
  const node = document.createElement("div");
  node.className = "files-context-menu-divider";
  return node;
}

function openEntry(entry) {
  if (typeof filesActions.openEntry === "function") {
    filesActions.openEntry(entry);
  }
}

function openShareForEntry(entry) {
  if (!entry || entry.getAttribute("data-entry-type") !== "file") {
    return;
  }
  if (typeof filesActions.openShare === "function") {
    filesActions.openShare(
      String(entry.getAttribute("data-entry-id") || ""),
      String(entry.getAttribute("data-file-name") || ""),
    );
  }
}

function openMove(entries) {
  if (!entries.length) {
    return;
  }
  moveEntries.openDialog(entries);
}

function cutEntries(entries) {
  if (!entries.length) {
    return;
  }
  moveEntries.cutEntries(entries);
  if (window.Toast) {
    window.Toast.success("Cut " + entries.length + (entries.length === 1 ? " item." : " items."), {
      title: "Ready to move"
    });
  }
}

async function pasteEntries(targetFolderId) {
  try {
    await moveEntries.pasteInto(targetFolderId);
  } catch (error) {
    showAppError(error, {
      code: "validation_failed",
      message: "Paste failed.",
    });
  }
}

function deleteEntries(entries) {
  document.dispatchEvent(new CustomEvent("arkive:entries-delete-request", {
    detail: { selectedEntries: entries }
  }));
}

function renameEntries(entries, entry) {
  if (!entries.length) {
    return;
  }
  if (entries.length > 1) {
    showAppError(null, {
      code: "unknown_error",
      message: "Rename only supports one item at a time.",
    });
    return;
  }
  document.dispatchEvent(new CustomEvent("arkive:rename-request", {
    detail: { entry: entry || null }
  }));
}

function resolveTargetSelection(entry) {
  const api = selectionAPI();
  const current = selectedEntries();
  const id = String(entry && entry.getAttribute("data-entry-id") || "");
  const inSelection = current.some(function(item) {
    return item && item.id === id;
  });
  if (entry && !inSelection && api && typeof api.selectOnly === "function") {
    api.selectOnly(entry);
  }
  return inSelection ? current : entry ? [{ id: id, type: String(entry.getAttribute("data-entry-type") || "") }] : current;
}

function uploadHere() {
  const link = document.querySelector(".files-upload-link");
  if (link && link.href) {
    window.location.href = link.href;
  }
}

function newFolderHere() {
  const button = document.getElementById("new-folder-button");
  if (button) {
    button.click();
  }
}

function selectAll() {
  const checkbox = document.getElementById("entries-select-all");
  if (checkbox && !checkbox.checked) {
    checkbox.checked = true;
    checkbox.dispatchEvent(new Event("change", { bubbles: true }));
  }
}

function currentFolderTargetId() {
  return moveEntries.currentTargetFolderId();
}

function canPasteTo(targetFolderId) {
  return moveEntries.canPasteTo(targetFolderId);
}

function hasClipboard() {
  return moveEntries.hasClipboard();
}

function clearCut() {
  moveEntries.clearClipboard();
  if (window.Toast) {
    window.Toast.success("Cut cancelled.", { title: "Clipboard cleared" });
  }
}

function entryMenuItems(entry, selection) {
  const type = String(entry.getAttribute("data-entry-type") || "");
  if (type === "folder") {
    return [
      { label: "Open", action: "open" },
      { label: "Cut", action: "cut" },
      { label: "Move...", action: "move" },
      { label: "Paste into folder", action: "paste-into-folder", disabled: !canPasteTo(entry.getAttribute("data-entry-id") || "") },
      { label: "Cancel cut", action: "clear-cut", disabled: !hasClipboard() },
      "divider",
      { label: "Rename", action: "rename" },
      { label: "Delete", action: "delete" },
    ];
  }
  return [
    { label: "Open", action: "open" },
    { label: "Share", action: "share" },
    { label: "Cut", action: "cut" },
    { label: "Move...", action: "move" },
    { label: "Rename", action: "rename" },
    { label: "Cancel cut", action: "clear-cut", disabled: !hasClipboard() },
    "divider",
    { label: "Delete", action: "delete" },
  ];
}

function emptyMenuItems() {
  return [
    { label: "Paste here", action: "paste-here", disabled: !canPasteTo(currentFolderTargetId()) },
    { label: "Cancel cut", action: "clear-cut", disabled: !hasClipboard() },
    { label: "New folder", action: "new-folder" },
    { label: "Upload here", action: "upload" },
    { label: "Select all", action: "select-all" },
  ];
}

function handleAction(action) {
  const state = menuState();
  const entry = state.currentEntry;
  const selection = state.currentSelection;

  switch (action) {
    case "open":
      openEntry(entry);
      return;
    case "share":
      openShareForEntry(entry);
      return;
    case "cut":
      cutEntries(selection);
      return;
    case "move":
      openMove(selection);
      return;
    case "paste-here":
      pasteEntries(currentFolderTargetId());
      return;
    case "paste-into-folder":
      pasteEntries(entry ? entry.getAttribute("data-entry-id") || "" : "");
      return;
    case "clear-cut":
      clearCut();
      return;
    case "delete":
      deleteEntries(selection);
      return;
    case "rename":
      renameEntries(selection, entry);
      return;
    case "new-folder":
      newFolderHere();
      return;
    case "upload":
      uploadHere();
      return;
    case "select-all":
      selectAll();
      return;
    default:
      return;
  }
}

function positionMenu(root, x, y) {
  root.style.left = "0px";
  root.style.top = "0px";
  root.hidden = false;
  const rect = root.getBoundingClientRect();
  const maxX = Math.max(8, window.innerWidth - rect.width - 8);
  const maxY = Math.max(8, window.innerHeight - rect.height - 8);
  root.style.left = Math.min(Math.max(8, x), maxX) + "px";
  root.style.top = Math.min(Math.max(8, y), maxY) + "px";
}

function showMenu(items, x, y, entry, selection) {
  const root = ensureMenu();
  root.innerHTML = "";
  items.forEach(function(item) {
    if (item === "divider") {
      root.appendChild(menuDivider());
      return;
    }
    root.appendChild(menuButton(item, handleAction));
  });
  const state = menuState();
  state.currentEntry = entry;
  state.currentSelection = selection;
  positionMenu(root, x, y);
  const firstButton = root.querySelector(".files-context-menu-item:not(:disabled)");
  if (firstButton) {
    firstButton.focus();
  }
}

export function initContextMenu() {
  ensureMenu();

  document.addEventListener("arkive:context-menu-close", closeMenu);

  document.addEventListener("click", function(event) {
    const trigger = event.target.closest("[data-entry-menu-trigger='true']");
    if (!trigger) {
      if (!event.target.closest(".files-context-menu")) {
        closeMenu();
      }
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    const entry = trigger.closest("[data-selectable-entry]");
    if (!entry) {
      return;
    }
    const selection = resolveTargetSelection(entry);
    const rect = trigger.getBoundingClientRect();
    showMenu(entryMenuItems(entry, selection), rect.right - 12, rect.bottom + 8, entry, selection);
  });

  document.addEventListener("contextmenu", function(event) {
    const entry = event.target && event.target.closest ? event.target.closest("[data-selectable-entry]") : null;
    const menu = event.target && event.target.closest ? event.target.closest(".files-context-menu") : null;
    if (menu) {
      return;
    }
    if (entry) {
      event.preventDefault();
      const selection = resolveTargetSelection(entry);
      showMenu(entryMenuItems(entry, selection), event.clientX, event.clientY, entry, selection);
      return;
    }
    const emptyArea = event.target && event.target.closest
      ? event.target.closest(".files-grid-wrap, .files-table-wrap")
      : null;
    if (!emptyArea) {
      closeMenu();
      return;
    }
    event.preventDefault();
    showMenu(emptyMenuItems(), event.clientX, event.clientY, null, selectedEntries());
  });

  document.addEventListener("scroll", closeMenu, true);
  window.addEventListener("resize", closeMenu);
}
