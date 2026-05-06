import { initCopyButtons } from "./features/copy_button.js";
import { initCrypto } from "./features/crypto.js";
import { initDialogs } from "./features/dialog.js";
import { initDropdowns } from "./features/dropdown.js";
import { initLightbox } from "./features/lightbox.js";
import { initSearchPalette } from "./features/search_palette.js";
import { initSidebar } from "./features/sidebar.js";
import { initToast } from "./features/toast.js";
import { initUploads } from "./features/uploads.js";
import { initVault } from "./features/vault.js";

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

function initRateLimitFetch() {
  if (!window.fetch || window.__arkiveRateLimitFetchReady) {
    return;
  }

  const originalFetch = window.fetch;
  let lastToastAt = 0;
  const rateLimitState = window.RateLimit || {};
  rateLimitState.until = rateLimitState.until || 0;
  rateLimitState.isActive = function() {
    return Date.now() < rateLimitState.until;
  };
  rateLimitState.setLimited = function(seconds) {
    const waitSeconds = typeof seconds === "number" && seconds > 0 ? seconds : 60;
    const now = Date.now();
    rateLimitState.until = Math.max(rateLimitState.until, now + waitSeconds * 1000);
  };
  window.RateLimit = rateLimitState;

  window.fetch = async function() {
    const res = await originalFetch.apply(this, arguments);
    if (!res || res.status !== 429) {
      return res;
    }
    const retryHeader = res.headers ? res.headers.get("Retry-After") : null;
    const limitedHeader = res.headers ? res.headers.get("X-Rate-Limited") : null;
    if (!retryHeader && !limitedHeader) {
      return res;
    }
    let retrySeconds = retryHeader ? parseInt(retryHeader, 10) : NaN;
    if (!isFinite(retrySeconds) || retrySeconds <= 0) {
      retrySeconds = 60;
    }
    rateLimitState.setLimited(retrySeconds);
    const now = Date.now();
    if (window.Toast && now - lastToastAt > 30000) {
      window.Toast.warning("Too many requests. Try again in a minute.", { title: "Slow down" });
      lastToastAt = now;
    }
    const err = new Error("Rate limited");
    err.status = 429;
    err.rateLimited = true;
    err.retryAfter = retrySeconds;
    throw err;
  };

  window.__arkiveRateLimitFetchReady = true;
}

initTheme();
initPasswordToggles();
initCrypto();
initVault();
initToast();
initRateLimitFetch();
initDialogs();
initDropdowns();
initCopyButtons();
initSidebar();
initSearchPalette();
initLightbox();
initUploads();
