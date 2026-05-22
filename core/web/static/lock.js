import { apiRequest } from "./lib/api.js";
import { getVaultState, persistSessionUnlock, unlockVault, waitUntilReady } from "./features/vault.js";

(function() {
  const form = document.querySelector("[data-lock-form='true']");
  if (!form) {
    return;
  }

  function lockRedirectTarget() {
    const raw = String(form.getAttribute("data-lock-next") || "").trim();
    if (!raw) {
      return "/dashboard";
    }
    if (raw.charAt(0) !== "/" || raw.indexOf("//") === 0) {
      return "/dashboard";
    }
    return raw;
  }

  function setGeneralError(message) {
    let error = form.querySelector(".form-error");
    if (!message) {
      if (error && error.hasAttribute("data-runtime-error")) {
        error.parentNode.removeChild(error);
      }
      return;
    }
    if (!error) {
      error = document.createElement("p");
      error.className = "form-error";
      error.setAttribute("data-runtime-error", "true");
      form.insertBefore(error, form.firstChild);
    }
    error.textContent = message;
  }

  async function apiUnlock(password) {
    return apiRequest("/api/auth/unlock", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ password: password })
    }, {
      code: "unauthorized",
      message: "Unlock failed. Try again."
    });
  }

  let submitting = false;

  waitUntilReady().then(function() {
    if (getVaultState().unlocked) {
      window.location.replace(lockRedirectTarget());
    }
  }).catch(function() {});

  form.addEventListener("submit", function(event) {
    event.preventDefault();
    if (submitting) {
      return;
    }

    const passwordInput = form.querySelector("input[name='password']");
    const submitButton = form.querySelector("button[type='submit']");
    const password = passwordInput ? String(passwordInput.value || "") : "";

    submitting = true;
    setGeneralError("");
    if (submitButton) {
      submitButton.disabled = true;
    }

    apiUnlock(password)
      .then(function(data) {
        return unlockVault(password, data.salt, data.encryptedMasterKey)
          .then(async function() {
            await persistSessionUnlock();
            if (passwordInput) {
              passwordInput.value = "";
            }
            window.location.replace(lockRedirectTarget());
          });
      })
      .catch(function(error) {
        setGeneralError(error && error.message ? error.message : "Unlock failed. Try again.");
      })
      .finally(function() {
        submitting = false;
        if (submitButton) {
          submitButton.disabled = false;
        }
      });
  });
})();
