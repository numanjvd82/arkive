import { apiRequest } from "./lib/api.js";
import { clearSessionUnlock, persistSessionUnlock, unlockVault } from "./features/vault.js";

(function() {
  const form = document.querySelector("[data-login-form='true']");
  if (!form) {
    return;
  }

  let submitting = false;

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

  async function apiLogin(email, password) {
    return apiRequest("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: email,
        password: password
      })
    }, {
      code: "unauthorized",
      message: "Unlock failed. Try again."
    });
  }

  async function rollbackSession() {
    try {
      await fetch("/logout", {
        method: "POST"
      });
    } catch (_) {}
  }

  form.addEventListener("submit", function(event) {
    event.preventDefault();
    if (submitting) {
      return;
    }

    const emailInput = form.querySelector("input[name='email']");
    const passwordInput = form.querySelector("input[name='password']");
    const submitButton = form.querySelector("button[type='submit']");
    const email = emailInput ? String(emailInput.value || "").trim() : "";
    const password = passwordInput ? String(passwordInput.value || "") : "";

    submitting = true;
    setGeneralError("");
    if (submitButton) {
      submitButton.disabled = true;
    }

    apiLogin(email, password)
      .then(function(data) {
        return unlockVault(password, data.salt, data.encryptedMasterKey)
          .then(async function() {
            await persistSessionUnlock();
            if (passwordInput) {
              passwordInput.value = "";
            }
            window.location.href = data.redirectTo || "/dashboard";
          })
          .catch(function(error) {
            clearSessionUnlock();
            return rollbackSession().then(function() {
              throw error;
            });
          });
      })
      .catch(function(error) {
        console.error(error);
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
