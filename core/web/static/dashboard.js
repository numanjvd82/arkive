(function() {
  const rows = document.querySelectorAll("[data-activity-open]");
  if (!rows.length) {
    return;
  }

  function openRow(row) {
    const href = String(row.getAttribute("data-activity-open") || "");
    if (!href) {
      return;
    }
    window.location.href = href;
  }

  rows.forEach(function(row) {
    row.addEventListener("click", function(event) {
      if (event.target.closest("a, button, input, label")) {
        return;
      }
      openRow(row);
    });

    row.addEventListener("keydown", function(event) {
      if (event.key !== "Enter" && event.key !== " " && event.key !== "Spacebar") {
        return;
      }
      event.preventDefault();
      openRow(row);
    });
  });
})();
