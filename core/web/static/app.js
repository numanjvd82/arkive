import { initCopyButtons } from "./features/copy_button.js";
import { setButtonBusy } from "./features/button_state.js";
import { initCrypto } from "./features/crypto.js";
import { initDialogs } from "./features/dialog.js";
import { initDropdowns } from "./features/dropdown.js";
import { initFileSelection } from "./features/file_selection.js";
import { initDragSelect } from "./features/drag_select.js";
import { initEntryDragMove } from "./features/entry_drag_move.js";
import { initFileListHydrator } from "./features/file_list_hydrator.js";
import { initFolders } from "./features/folders.js";
import { initLightbox } from "./features/lightbox.js";
import { initMoveEntries } from "./features/move_entries.js";
import { initSearchPalette } from "./features/search_palette.js";
import { initRenameEntries } from "./features/rename_entries.js";
import { initSidebar } from "./features/sidebar.js";
import { initContextMenu } from "./features/context_menu.js";
import { initToast } from "./features/toast.js";
import { initTooltips } from "./features/tooltip.js";
import { initUploads } from "./features/uploads.js";
import { getVaultState, initVault, lockVault, onVaultLock, waitUntilReady } from "./features/vault.js";

function initTheme() {
  const modes = ["dark", "system", "light"];
  const stored = localStorage.getItem("theme");
  const current = modes.indexOf(stored) !== -1 ? stored : "system";
  const root = document.documentElement;
  const prefersLight = window.matchMedia && window.matchMedia("(prefers-color-scheme: light)");

  function updateToggle(mode) {
    const toggle = document.getElementById("theme-toggle");
    if (!toggle) {
      return;
    }
    const label = toggle.querySelector(".theme-label");
    if (label) {
      label.textContent = mode;
    } else {
      toggle.textContent = mode;
    }
    toggle.setAttribute("aria-label", "Theme: " + mode);
  }

  function resolvedTheme(mode) {
    if (mode !== "system") {
      return mode;
    }
    return prefersLight && prefersLight.matches ? "light" : "dark";
  }

  function setMode(mode) {
    root.setAttribute("data-theme", resolvedTheme(mode));
    root.setAttribute("data-theme-mode", mode);
    localStorage.setItem("theme", mode);
    updateToggle(mode);
  }

  setMode(current);

  if (prefersLight && prefersLight.addEventListener) {
    prefersLight.addEventListener("change", function() {
      if ((root.getAttribute("data-theme-mode") || "system") === "system") {
        setMode("system");
      }
    });
  }

  const toggle = document.getElementById("theme-toggle");
  if (!toggle || toggle.hasAttribute("data-theme-bound")) {
    return;
  }
  toggle.setAttribute("data-theme-bound", "true");
  toggle.addEventListener("click", function() {
    const active = root.getAttribute("data-theme-mode") || "system";
    const index = modes.indexOf(active);
    const next = modes[(index + 1) % modes.length];
    setMode(next);
  });
}

function initPasswordToggles() {
  const toggles = document.querySelectorAll(".password-toggle");
  if (!toggles.length) {
    return;
  }

  toggles.forEach(function(toggle) {
    if (toggle.hasAttribute("data-password-bound")) {
      return;
    }
    toggle.setAttribute("data-password-bound", "true");
    toggle.addEventListener("click", function() {
      const targetId = toggle.getAttribute("data-target");
      if (!targetId) {
        return;
      }
      const input = document.getElementById(targetId);
      if (!input) {
        return;
      }
      const visible = input.getAttribute("type") === "text";
      const nextVisible = !visible;
      input.setAttribute("type", nextVisible ? "text" : "password");
      toggle.setAttribute("aria-pressed", nextVisible ? "true" : "false");
      toggle.setAttribute("data-visible", nextVisible ? "true" : "false");
      toggle.setAttribute("aria-label", nextVisible ? "Hide password" : "Show password");
    });
  });
}

function initSubmitButtonStates() {
  document.addEventListener("submit", function(event) {
    const form = event.target;
    if (!form || !(form instanceof HTMLFormElement)) {
      return;
    }
    const submitter = event.submitter;
    if (!submitter) {
      return;
    }
    const busyText = String(submitter.getAttribute("data-busy-text") || "");
    setButtonBusy(submitter, true, { busyText: busyText });
  });
}

function buildLockURL() {
  const next = window.location.pathname + window.location.search + window.location.hash;
  return "/lock?next=" + encodeURIComponent(next || "/dashboard");
}

function initVaultAccessGuard() {
  const requiresUnlock = document.body && document.body.getAttribute("data-require-vault-unlock") === "true";
  const lockButton = document.getElementById("app-lock-trigger");

  function redirectToLock() {
    if (window.location.pathname === "/lock") {
      return;
    }
    window.location.replace(buildLockURL());
  }

  if (lockButton && !lockButton.hasAttribute("data-lock-bound")) {
    lockButton.setAttribute("data-lock-bound", "true");
    lockButton.addEventListener("click", function() {
      lockVault()
        .catch(function() {})
        .finally(function() {
          redirectToLock();
        });
    });
  }

  if (!requiresUnlock) {
    return;
  }

  waitUntilReady().then(function() {
    if (!getVaultState().unlocked) {
      redirectToLock();
    }
  }).catch(function() {
    redirectToLock();
  });

  onVaultLock(function() {
    redirectToLock();
  });
}

initTheme();
initPasswordToggles();
initSubmitButtonStates();
initCrypto();
initVault();
initToast();
initVaultAccessGuard();
initDialogs();
initDropdowns();
initTooltips();
initCopyButtons();
initSidebar();
initSearchPalette();
initLightbox();
initFileListHydrator();
initFileSelection();
initFolders();
initMoveEntries();
initRenameEntries();
initContextMenu();
initDragSelect();
initEntryDragMove();
initUploads();

if ("serviceWorker" in navigator && window.isSecureContext) {
  navigator.serviceWorker.register("/sw.js").catch(function(error) {
    console.warn("Service worker registration failed", error);
  });
}
