import { initCopyButtons } from "./features/copy_button.js";
import { initCrypto } from "./features/crypto.js";
import { initDialogs } from "./features/dialog.js";
import { initDropdowns } from "./features/dropdown.js";
import { ArkiveFileReader } from "./features/file_reader.js";
import * as ArkiveDownloadWarning from "./features/reader/download_warning.js";
import { initFileListHydrator } from "./features/file_list_hydrator.js";
import { initLightbox } from "./features/lightbox.js";
import { initSearchPalette } from "./features/search_palette.js";
import { ArkiveShareReader } from "./features/share_reader.js";
import { initSidebar } from "./features/sidebar.js";
import { initToast } from "./features/toast.js";
import { initTooltips } from "./features/tooltip.js";
import { initUploads } from "./features/uploads.js";
import { initVault } from "./features/vault.js";

window.ArkiveFileReader = ArkiveFileReader;
window.ArkiveShareReader = ArkiveShareReader;
window.ArkiveDownloadWarning = ArkiveDownloadWarning;

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
      if (!window.ArkiveVault || typeof window.ArkiveVault.lock !== "function") {
        redirectToLock();
        return;
      }
      window.ArkiveVault.lock()
        .catch(function() {})
        .finally(function() {
          redirectToLock();
        });
    });
  }

  if (!requiresUnlock || !window.ArkiveVault) {
    return;
  }

  if (typeof window.ArkiveVault.waitUntilReady === "function") {
    window.ArkiveVault.waitUntilReady().then(function() {
      if (typeof window.ArkiveVault.isUnlocked === "function" && !window.ArkiveVault.isUnlocked()) {
        redirectToLock();
      }
    }).catch(function() {
      redirectToLock();
    });
  }

  window.addEventListener("arkive:vault-state", function(event) {
    const detail = event && event.detail ? event.detail : {};
    if (!detail.unlocked) {
      redirectToLock();
    }
  });
}

initTheme();
initPasswordToggles();
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
initUploads();
