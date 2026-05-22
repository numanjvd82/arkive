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

    fetch(action, {
      method: "POST",
      body: body,
      headers: {
        "X-Requested-With": "XMLHttpRequest",
      },
    })
      .then(async function(res) {
        if (res.status === 204) {
          window.location.assign(action + window.location.hash);
          return null;
        }
        return window.ArkiveAPI.readJSON(res, "Unlock failed");
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
