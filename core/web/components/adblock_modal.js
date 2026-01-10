const BACKDROP_ID = "adblock-backdrop";
const STEPS_ID = "adblock-steps";
const HELP_ID = "adblock-help";
const CONTINUE_ID = "adblock-continue";
const CLOSE_ID = "adblock-close";
const WARNING_ID = "adblock-warning";
const DETECTION_DELAY_MS = 2500;
const RECHECK_DELAY_MS = 800;

function openAdblockModal() {
  if (!window.Dialog || !window.Dialog.open) {
    return;
  }
  window.Dialog.open(BACKDROP_ID);
  document.body.classList.add("adblock-locked");
}

function closeAdblockModal() {
  if (!window.Dialog || !window.Dialog.close) {
    return;
  }
  window.Dialog.close(BACKDROP_ID);
  document.body.classList.remove("adblock-locked");
}

function ensureBlockedWarning(isBlocked) {
  const warning = document.getElementById(WARNING_ID);
  if (!warning) {
    return;
  }
  if (isBlocked) {
    warning.removeAttribute("hidden");
  } else {
    warning.setAttribute("hidden", "hidden");
  }
}

function attemptCloseAdblockModal() {
  setContinueLabel(true);
  ensureBlockedWarning(true);
  setTimeout(function() {
    const stillBlocked = detectAdblock();
    setContinueLabel(stillBlocked);
    ensureBlockedWarning(stillBlocked);
    if (!stillBlocked) {
      closeAdblockModal();
    }
  }, RECHECK_DELAY_MS);
}

function toggleSteps() {
  const steps = document.getElementById(STEPS_ID);
  const helpButton = document.getElementById(HELP_ID);
  if (!steps || !helpButton) {
    return;
  }
  const isHidden = steps.hasAttribute("hidden");
  if (isHidden) {
    steps.removeAttribute("hidden");
    helpButton.setAttribute("aria-expanded", "true");
  } else {
    steps.setAttribute("hidden", "hidden");
    helpButton.setAttribute("aria-expanded", "false");
  }
}

function isBaitBlocked() {
  if (!document.body) {
    return false;
  }
  const bait = document.createElement("div");
  bait.className = "ad ads ad-banner ad-slot sponsored adsbox";
  bait.style.cssText = "position:absolute; left:-9999px; top:-9999px; width:1px; height:1px;";
  document.body.appendChild(bait);

  const style = window.getComputedStyle(bait);
  const blocked =
    style.display === "none" ||
    style.visibility === "hidden" ||
    bait.offsetHeight === 0 ||
    bait.offsetWidth === 0;

  bait.remove();
  return blocked;
}

function hasMonetagOnclick() {
  return Boolean(document.querySelector('script[src*="monetag-onclick.js"]'));
}

function hasMonetagVignette() {
  return Boolean(document.querySelector('script[src*="monetag-vignette.js"]'));
}

function hasMonetagPush() {
  return Boolean(document.querySelector('script[src*="monetag-push-ad.js"]'));
}

function hasMonetagOnclickInjection() {
  return Boolean(
    document.querySelector('script[data-zone="10431804"], script[src*="al5sm.com/tag.min.js"]')
  );
}

function hasMonetagVignetteInjection() {
  return Boolean(
    document.querySelector('script[data-zone="10431810"], script[src*="gizokraijaw.net/vignette.min.js"]')
  );
}

function hasMonetagPushInjection() {
  return Boolean(
    document.querySelector('script[src*="3nbf4.com/act/files/tag.min.js"]')
  );
}

function hasMonetagResourceLoaded() {
  if (!window.performance || !window.performance.getEntriesByType) {
    return false;
  }

  const resources = window.performance.getEntriesByType("resource");
  return resources.some(function(entry) {
    if (!entry || !entry.name) {
      return false;
    }
    return (
      entry.name.includes("al5sm.com/") ||
      entry.name.includes("gizokraijaw.net/") ||
      entry.name.includes("3nbf4.com/act/files/tag.min.js")
    );
  });
}

function isAdsterraBlocked() {
  const adsterraExpected = document.querySelector(
    'script[src*="effectivegatecpm.com"][src*="invoke.js"]'
  );
  if (!adsterraExpected) {
    return false;
  }

  const container = document.getElementById("container-3e709d756892597be3b0708e86694b25");
  if (!container) {
    return false;
  }

  const hasChildren = container.children.length > 0;
  return !hasChildren;
}

function detectAdblock() {
  const baitBlocked = isBaitBlocked();
  const monetagBlocked =
    (hasMonetagOnclick() && !hasMonetagOnclickInjection() && !hasMonetagResourceLoaded()) ||
    (hasMonetagVignette() && !hasMonetagVignetteInjection() && !hasMonetagResourceLoaded()) ||
    (hasMonetagPush() && !hasMonetagPushInjection() && !hasMonetagResourceLoaded());
  const adsterraBlocked = isAdsterraBlocked();

  return baitBlocked || monetagBlocked || adsterraBlocked;
}

function setupAdblockModal() {
  const helpButton = document.getElementById(HELP_ID);
  const continueButton = document.getElementById(CONTINUE_ID);
  const closeButton = document.getElementById(CLOSE_ID);
  const backdrop = document.getElementById(BACKDROP_ID);

  if (helpButton) {
    helpButton.setAttribute("aria-controls", STEPS_ID);
    helpButton.setAttribute("aria-expanded", "false");
    helpButton.addEventListener("click", toggleSteps);
  }

  if (continueButton) {
    continueButton.addEventListener("click", attemptCloseAdblockModal);
  }

  if (closeButton) {
    closeButton.addEventListener("click", attemptCloseAdblockModal);
  }

  if (backdrop) {
    backdrop.addEventListener("click", function(event) {
      event.stopPropagation();
      attemptCloseAdblockModal();
    });
  }

  document.addEventListener(
    "keydown",
    function(event) {
      if (event.key !== "Escape") {
        return;
      }
      const isOpen = backdrop && !backdrop.classList.contains("is-hidden");
      if (!isOpen) {
        return;
      }
      event.preventDefault();
      event.stopPropagation();
      attemptCloseAdblockModal();
    },
    true
  );
}

function setContinueLabel(isBlocked) {
  const continueButton = document.getElementById(CONTINUE_ID);
  if (!continueButton) {
    return;
  }
  continueButton.textContent = isBlocked ? "Check again" : "Continue";
}

// Optional: block download button until adblock is disabled.
// function blockDownloadButton() {
//   const button = document.getElementById("downloadBtn");
//   if (!button) {
//     return;
//   }
//   button.setAttribute("disabled", "disabled");
//   button.classList.add("is-disabled");
// }

setTimeout(function() {
  if (!document.getElementById(BACKDROP_ID)) {
    return;
  }
  setupAdblockModal();

  const firstPass = detectAdblock();
  if (!firstPass) {
    ensureBlockedWarning(false);
    setContinueLabel(false);
    return;
  }

  setTimeout(function() {
    const secondPass = detectAdblock();
    if (secondPass) {
      ensureBlockedWarning(true);
      setContinueLabel(true);
      openAdblockModal();
      // blockDownloadButton();
    } else {
      ensureBlockedWarning(false);
      setContinueLabel(false);
    }
  }, RECHECK_DELAY_MS);
}, DETECTION_DELAY_MS);
