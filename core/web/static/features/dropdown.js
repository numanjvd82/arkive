export function initDropdowns() {
  if (document.documentElement.hasAttribute("data-dropdown-ready")) {
    return;
  }
  document.documentElement.setAttribute("data-dropdown-ready", "true");

  function setExpanded(dropdown, isOpen) {
    const trigger = dropdown.querySelector("[data-dropdown-trigger]");
    if (trigger) {
      trigger.setAttribute("aria-expanded", isOpen ? "true" : "false");
    }
  }

  function setRowOpenState(dropdown, isOpen) {
    const row = dropdown.closest(".files-row");
    if (!row) {
      return;
    }
    if (isOpen) {
      row.classList.add("is-dropdown-open");
    } else {
      row.classList.remove("is-dropdown-open");
    }
  }

  function closeDropdown(dropdown) {
    if (!dropdown) {
      return;
    }
    dropdown.classList.remove("is-open");
    setExpanded(dropdown, false);
    setRowOpenState(dropdown, false);
  }

  function closeAll(except) {
    document.querySelectorAll("[data-dropdown].is-open").forEach(function(dropdown) {
      if (dropdown !== except) {
        closeDropdown(dropdown);
      }
    });
  }

  document.addEventListener("click", function(event) {
    const trigger = event.target.closest("[data-dropdown-trigger]");
    if (trigger) {
      event.preventDefault();
      event.stopPropagation();
      const dropdown = trigger.closest("[data-dropdown]");
      if (!dropdown) {
        return;
      }
      const isOpen = dropdown.classList.contains("is-open");
      if (isOpen) {
        closeDropdown(dropdown);
      } else {
        closeAll(dropdown);
        dropdown.classList.add("is-open");
        setExpanded(dropdown, true);
        setRowOpenState(dropdown, true);
      }
      return;
    }

    const menu = event.target.closest("[data-dropdown-menu]");
    if (menu) {
      const dropdown = menu.closest("[data-dropdown]");
      if (!dropdown) {
        return;
      }
      if (event.target.closest("a, button")) {
        closeDropdown(dropdown);
      }
      return;
    }

    closeAll(null);
  });

  document.addEventListener("keydown", function(event) {
    if (event.key !== "Escape") {
      return;
    }
    closeAll(null);
  });
}
