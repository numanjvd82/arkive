import { apiRequest } from "../lib/api.js";
import { clearSessionUnlock } from "./vault.js";

(function() {
  const forgotForm = document.querySelector("[data-forgot-password-form='true']");
  const resetForm = document.querySelector("[data-reset-password-form='true']");
  let worker = null;
  let workerReady = null;

  function setMessage(node, message, isError) {
    if (!node) {
      return;
    }
    node.hidden = !message;
    node.textContent = message || "";
    node.classList.toggle("form-error", !!isError);
  }

  async function ensureWorker() {
    if (workerReady) {
      return workerReady;
    }
    worker = new Worker("/static/workers/crypto_worker.js", { type: "module" });
    workerReady = Promise.resolve(worker);
    return workerReady;
  }

  async function callWorker(method, params) {
    const activeWorker = await ensureWorker();
    return new Promise(function(resolve, reject) {
      const id = Math.random().toString(36).slice(2);
      function onMessage(event) {
        const data = event.data || {};
        if (data.id !== id) {
          return;
        }
        activeWorker.removeEventListener("message", onMessage);
        if (data.ok) {
          resolve(data.result || {});
          return;
        }
        reject(new Error(data.error || "Password reset failed"));
      }
      activeWorker.addEventListener("message", onMessage);
      activeWorker.postMessage({ id: id, method: method, params: params || {} });
    });
  }

  if (forgotForm) {
    const resultNode = forgotForm.querySelector("[data-forgot-password-result]");
    forgotForm.addEventListener("submit", function(event) {
      event.preventDefault();
      const email = forgotForm.querySelector("input[name='email']");
      const submit = forgotForm.querySelector("button[type='submit']");
      if (submit) {
        submit.disabled = true;
      }
      setMessage(resultNode, "", false);
      apiRequest("/api/auth/forgot-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email ? email.value : "" }),
      }, {
        message: "Password reset failed.",
      })
        .then(function(data) {
          const url = data && data.resetURL ? data.resetURL : "";
          if (url) {
            // Core returns `resetURL` directly (no mailer). Redirect for convenience.
            try {
              const target = new URL(String(url), window.location.origin);
              window.location.assign(target.toString());
              return;
            } catch (_) {
              // Fall back to showing the URL if parsing fails.
            }
          }
          setMessage(resultNode, url ? "Reset link: " + url : "Reset link created.", false);
        })
        .catch(function(error) {
          console.error(error);
          setMessage(resultNode, error && error.message ? error.message : "Password reset failed.", true);
        })
        .finally(function() {
          if (submit) {
            submit.disabled = false;
          }
        });
    });
  }

  if (resetForm) {
    const resultNode = resetForm.querySelector("[data-reset-password-result]");
    resetForm.addEventListener("submit", function(event) {
      event.preventDefault();
      const token = new URLSearchParams(window.location.search).get("token") || "";
      const recoveryKey = resetForm.querySelector("input[name='recovery_key']");
      const newPassword = resetForm.querySelector("input[name='new_password']");
      const confirmPassword = resetForm.querySelector("input[name='confirm_password']");
      const submit = resetForm.querySelector("button[type='submit']");

      if (!token) {
        setMessage(resultNode, "Reset token is missing.", true);
        return;
      }
      if (!newPassword || !confirmPassword || newPassword.value !== confirmPassword.value) {
        setMessage(resultNode, "Passwords do not match.", true);
        return;
      }
      if (submit) {
        submit.disabled = true;
      }
      setMessage(resultNode, "", false);

      apiRequest("/api/auth/password-reset/vault?token=" + encodeURIComponent(token), {
        method: "GET",
      }, {
        message: "Password reset failed.",
      })
        .then(function(vaultData) {
          return callWorker("preparePasswordReset", {
            encryptedMasterKeyRecovery: vaultData.encryptedMasterKeyRecovery,
            recoveryKey: recoveryKey ? recoveryKey.value : "",
            newPassword: newPassword ? newPassword.value : "",
            userID: vaultData.userID,
          });
        })
        .then(function(material) {
          return apiRequest("/api/auth/password-reset/complete", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              token: token,
              newPassword: newPassword ? newPassword.value : "",
              vaultSalt: material.vaultSalt,
              encryptedMasterKey: material.encryptedMasterKey,
            }),
          }, {
            message: "Password reset failed.",
          });
        })
        .then(function(data) {
          clearSessionUnlock();
          sessionStorage.clear();
          window.location.href = (data && data.redirectTo) || "/login";
        })
        .catch(function(error) {
          console.error(error);
          setMessage(resultNode, error && error.message ? error.message : "Password reset failed.", true);
        })
        .finally(function() {
          if (submit) {
            submit.disabled = false;
          }
        });
    });
  }
})();
