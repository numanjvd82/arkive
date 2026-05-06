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

  function firstError(errors) {
    if (!errors) {
      return "";
    }
    if (typeof errors.general === "string" && errors.general) {
      return errors.general;
    }
    if (typeof errors._general === "string" && errors._general) {
      return errors._general;
    }
    const keys = Object.keys(errors);
    for (let i = 0; i < keys.length; i++) {
      const value = errors[keys[i]];
      if (typeof value === "string" && value) {
        return value;
      }
    }
    return "";
  }

  async function apiLogin(email, password) {
    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: email,
        password: password
      })
    });
    const data = await response.json().catch(function() {
      return {};
    });
    if (!response.ok) {
      const message = firstError(data.errors) || data.error || "Unlock failed. Try again.";
      throw new Error(message);
    }
    return data;
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

    if (!window.ArkiveVault || typeof window.ArkiveVault.unlockVault !== "function") {
      setGeneralError("Vault runtime is unavailable. Reload the page and try again.");
      return;
    }

    submitting = true;
    setGeneralError("");
    if (submitButton) {
      submitButton.disabled = true;
    }

    apiLogin(email, password)
      .then(function(data) {
        return window.ArkiveVault.unlockVault(password, data.salt, data.encryptedMasterKey)
          .then(function() {
            if (passwordInput) {
              passwordInput.value = "";
            }
            window.location.href = data.redirectTo || "/dashboard";
          })
          .catch(function(error) {
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
