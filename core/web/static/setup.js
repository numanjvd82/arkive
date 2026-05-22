import { getArkiveCrypto } from "./features/crypto.js";

(function() {
  const form = document.querySelector("[data-setup-vault-form='true']");
  const saltInput = document.querySelector("[data-vault-salt-input='true']");
  const encryptedMasterKeyInput = document.querySelector("[data-encrypted-master-key-input='true']");

  if (!form || !saltInput || !encryptedMasterKeyInput) {
    return;
  }

  let submitting = false;

  function bytesToBase64(bytes) {
    let binary = "";
    for (let i = 0; i < bytes.length; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }

  function setFormError(message) {
    const body = form.querySelector(".setup-body");
    if (!body) {
      return;
    }
    let error = body.querySelector(".form-error");
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
      body.insertBefore(error, body.firstChild);
    }
    error.textContent = message;
  }

  async function provisionVaultMaterial(password) {
    const crypto = await getArkiveCrypto();
    const salt = crypto.generate_salt();
    const masterKey = crypto.generate_master_key();
    const kek = crypto.derive_password_kek(password, salt);
    const aad = new TextEncoder().encode("arkive:master-key:v1");
    const encryptedMasterKey = crypto.wrap_master_key(masterKey, kek, aad);

    try {
      saltInput.value = bytesToBase64(salt);
      encryptedMasterKeyInput.value = bytesToBase64(encryptedMasterKey);
    } finally {
      crypto.zeroize(salt);
      crypto.zeroize(masterKey);
      crypto.zeroize(kek);
      crypto.zeroize(encryptedMasterKey);
    }
  }

  form.addEventListener("submit", function(event) {
    if (submitting) {
      return;
    }
    if (saltInput.value && encryptedMasterKeyInput.value) {
      return;
    }

    event.preventDefault();
    const passwordInput = form.querySelector("input[name='password']");
    const password = passwordInput ? String(passwordInput.value || "") : "";
    submitting = true;
    setFormError("");

    provisionVaultMaterial(password)
      .then(function() {
        HTMLFormElement.prototype.submit.call(form);
      })
      .catch(function(error) {
        console.error(error);
        submitting = false;
        setFormError("Vault bootstrap failed. Reload the page and try again.");
      });
  });
})();
