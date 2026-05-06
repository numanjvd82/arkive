export function initSearchPalette() {
  const trigger = document.getElementById("app-search-trigger");
  const panel = document.getElementById("search-panel");
  const input = document.getElementById("search-panel-input");
  const filesRoot = document.getElementById("search-results-files");
  const sharesRoot = document.getElementById("search-results-shares");
  const settingsRoot = document.getElementById("search-results-settings");
  const empty = document.getElementById("search-empty");

  if (!trigger || !panel || !input || !filesRoot || !sharesRoot || !settingsRoot || !empty || trigger.hasAttribute("data-search-bound")) {
    return;
  }
  trigger.setAttribute("data-search-bound", "true");

  const sections = {
    files: document.querySelector("[data-search-category='files']"),
    shares: document.querySelector("[data-search-category='shares']"),
    settings: document.querySelector("[data-search-category='settings']")
  };

  let activeIndex = 0;
  let flatItems = [];
  let abortController = null;
  let searchTimer = null;

  function iconFor(kind) {
    if (kind === "file") return "F";
    if (kind === "share") return "S";
    return "C";
  }

  function openPanel(seed) {
    panel.classList.remove("is-hidden");
    panel.setAttribute("aria-hidden", "false");
    if (typeof seed === "string") {
      input.value = seed;
    }
    window.requestAnimationFrame(function() {
      input.focus();
      input.setSelectionRange(input.value.length, input.value.length);
    });
  }

  function closePanel() {
    panel.classList.add("is-hidden");
    panel.setAttribute("aria-hidden", "true");
    flatItems = [];
    activeIndex = 0;
  }

  function syncActive() {
    const items = panel.querySelectorAll(".search-item");
    items.forEach(function(item, index) {
      item.classList.toggle("is-active", index === activeIndex);
    });
  }

  function flatten() {
    flatItems = Array.from(panel.querySelectorAll(".search-item"));
    if (!flatItems.length) {
      activeIndex = 0;
      return;
    }
    if (activeIndex >= flatItems.length) {
      activeIndex = 0;
    }
    syncActive();
  }

  function sectionVisibility(key, count) {
    if (!sections[key]) {
      return;
    }
    sections[key].classList.toggle("is-hidden", count === 0);
  }

  function makeItem(result) {
    const item = document.createElement("a");
    item.className = "search-item";
    item.href = result.url;
    item.dataset.url = result.url;

    const main = document.createElement("div");
    main.className = "search-item-main";

    const iconWrap = document.createElement("div");
    iconWrap.className = "search-item-icon";
    iconWrap.textContent = iconFor(result.kind);

    const copy = document.createElement("div");
    copy.className = "search-item-copy";

    const title = document.createElement("div");
    title.className = "search-item-title";
    title.textContent = result.title;
    copy.appendChild(title);

    if (result.meta) {
      const meta = document.createElement("div");
      meta.className = "search-item-meta";
      meta.textContent = result.meta;
      copy.appendChild(meta);
    }

    main.appendChild(iconWrap);
    main.appendChild(copy);
    item.appendChild(main);

    if (result.status) {
      const status = document.createElement("div");
      status.className = "search-item-status";
      status.textContent = result.status;
      item.appendChild(status);
    }

    return item;
  }

  function renderCategory(root, key, items) {
    root.innerHTML = "";
    items.forEach(function(result) {
      root.appendChild(makeItem(result));
    });
    sectionVisibility(key, items.length);
  }

  function renderResults(results) {
    const files = (results && results.files) || [];
    const shares = (results && results.shares) || [];
    const settings = (results && results.settings) || [];

    renderCategory(filesRoot, "files", files);
    renderCategory(sharesRoot, "shares", shares);
    renderCategory(settingsRoot, "settings", settings);

    const total = files.length + shares.length + settings.length;
    empty.classList.toggle("is-hidden", total !== 0);
    flatten();
  }

  function runSearch(query) {
    if (abortController) {
      abortController.abort();
    }
    abortController = new AbortController();
    fetch("/api/search?q=" + encodeURIComponent(query), {
      method: "GET",
      signal: abortController.signal,
      headers: { "Accept": "application/json" }
    })
      .then(function(res) {
        if (!res.ok) throw new Error("search failed");
        return res.json();
      })
      .then(function(data) {
        renderResults(data.results || {});
      })
      .catch(function(err) {
        if (err && err.name === "AbortError") return;
        renderResults({});
      });
  }

  function scheduleSearch() {
    const query = input.value.trim();
    if (searchTimer) {
      clearTimeout(searchTimer);
    }
    if (!query) {
      renderResults({});
      return;
    }
    searchTimer = setTimeout(function() {
      runSearch(query);
    }, 120);
  }

  trigger.addEventListener("click", function() {
    openPanel("");
  });

  trigger.addEventListener("keydown", function(event) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      openPanel("");
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      openPanel("");
      if (!flatItems.length) return;
      activeIndex = 0;
      syncActive();
    }
  });

  input.addEventListener("input", function() {
    scheduleSearch();
  });

  input.addEventListener("keydown", function(event) {
    if (event.key === "Escape") {
      event.preventDefault();
      closePanel();
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      if (!flatItems.length) return;
      activeIndex = (activeIndex + 1) % flatItems.length;
      syncActive();
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      if (!flatItems.length) return;
      activeIndex = (activeIndex - 1 + flatItems.length) % flatItems.length;
      syncActive();
      return;
    }
    if (event.key === "Enter") {
      if (!flatItems.length) return;
      event.preventDefault();
      window.location.href = flatItems[activeIndex].dataset.url;
    }
  });

  panel.addEventListener("mousedown", function(event) {
    event.stopPropagation();
  });

  panel.addEventListener("click", function(event) {
    const item = event.target.closest(".search-item");
    if (!item) {
      return;
    }
    closePanel();
  });

  document.addEventListener("click", function(event) {
    if (panel.classList.contains("is-hidden")) return;
    if (event.target === trigger || trigger.contains(event.target)) return;
    if (panel.contains(event.target) && event.target !== panel) return;
    closePanel();
  });

  document.addEventListener("keydown", function(event) {
    if (event.key === "Escape" && !panel.classList.contains("is-hidden")) {
      event.preventDefault();
      closePanel();
      return;
    }
    if (event.key === "/" && document.activeElement !== input && document.activeElement !== trigger) {
      const tag = document.activeElement && document.activeElement.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA") return;
      event.preventDefault();
      openPanel("");
      scheduleSearch();
    }
  });

  window.addEventListener("hashchange", closePanel);
  window.addEventListener("popstate", closePanel);
}
