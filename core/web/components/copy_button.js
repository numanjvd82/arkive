(function() {
  const buttons = document.querySelectorAll("[data-copy-button]");
  if (!buttons.length) {
    return;
  }

  function getCopyValue(button) {
    const value = button.getAttribute("data-copy-value");
    if (value) {
      return value;
    }
    const targetId = button.getAttribute("data-copy-target");
    if (!targetId) {
      return "";
    }
    const target = document.getElementById(targetId);
    if (!target) {
      return "";
    }
    if (target.value !== undefined) {
      return target.value;
    }
    return target.textContent || "";
  }

  function writeToClipboard(text) {
    if (!text) {
      return Promise.reject();
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text);
    }
    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.setAttribute("readonly", "readonly");
    textarea.style.position = "absolute";
    textarea.style.left = "-9999px";
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
    return Promise.resolve();
  }

  buttons.forEach(function(button) {
    button.addEventListener("click", function() {
      const value = getCopyValue(button);
      const successTitle = button.getAttribute("data-copy-success-title") || "Copied";
      const successMessage = button.getAttribute("data-copy-success-message") || "Link copied.";
      writeToClipboard(value)
        .then(function() {
          if (window.Toast) {
            window.Toast.success(successMessage, { title: successTitle });
          }
        })
        .catch(function() {
          if (window.Toast) {
            window.Toast.error("Copy failed. Try again.");
          }
        });
    });
  });
})();
