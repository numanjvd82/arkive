import { apiRequest } from "./lib/api.js";

(function() {
  const form = document.querySelector(".share-form");
  const errorEl = document.querySelector(".form-error");
  const submit = document.querySelector(".share-submit");

  if (!form) {
    return;
  }

  form.addEventListener("submit", function(event) {
    event.preventDefault();
    const action = form.getAttribute("action") || window.location.pathname;
    const body = new FormData(form);
    if (errorEl) {
      errorEl.textContent = "";
    }
    if (submit) {
      submit.disabled = true;
    }

    apiRequest(action, {
      method: "POST",
      body: body,
      headers: {
        "X-Requested-With": "XMLHttpRequest",
      },
    }, {
      code: "forbidden",
      message: "Unlock failed",
    })
      .then(function() {
        window.location.assign(action + window.location.hash);
      })
      .catch(function(error) {
        if (errorEl) {
          errorEl.textContent = (error && error.message) || "Unlock failed";
        }
      })
      .finally(function() {
        if (submit) {
          submit.disabled = false;
        }
      });
  });
})();
