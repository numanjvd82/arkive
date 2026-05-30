import { apiRequest } from "../lib/api.js";
import { thumbnailCache } from "../upload/thumbnail_cache.js";
import { vault } from "./vault.js";

const SEARCH_DEBOUNCE_MS = 300;
const MIN_SEARCH_QUERY_LEN = 3;

const SETTINGS = [
  { id: "settings-0", title: "Instance Overview", terms: ["instance overview", "instance", "admin", "email", "storage", "usage"], url: "/settings#settings-account", meta: "Settings" },
  { id: "settings-1", title: "Storage Provider", terms: ["storage provider", "storage configuration", "provider", "local", "s3"], url: "/settings#settings-provider", meta: "Settings" },
  { id: "settings-2", title: "Security", terms: ["security", "authentication", "session", "hardening"], url: "/settings#settings-security", meta: "Settings" },
];

function normalizeQuery(value) {
  return String(value || "")
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[^\p{L}\p{N}.]+/gu, " ")
    .trim();
}

function searchSettingsResults(query) {
  const normalized = normalizeQuery(query);
  if (!normalized) {
    return [];
  }
  return SETTINGS.filter(function(item) {
    return item.terms.some(function(term) {
      return normalizeQuery(term).includes(normalized);
    });
  }).map(function(item) {
    return {
      id: item.id,
      kind: "setting",
      title: item.title,
      meta: item.meta,
      url: item.url,
      category: "Settings",
    };
  });
}

async function decryptSearchFiles(files) {
  const results = [];
  for (let i = 0; i < files.length; i += 1) {
    const file = files[i];
    try {
      const opened = await vault.decryptSearchFile(file);
      const metadata = opened && opened.metadata ? opened.metadata : {};
      results.push({
        id: file.id,
        kind: "file",
        title: String(metadata.name || "Encrypted file"),
        meta: String(metadata.mime || "Encrypted"),
        score: Number(file.score || 0),
        preview: metadata && metadata.preview ? metadata.preview : null,
        encryptedFileKey: file.encryptedFileKey,
        vaultId: file.vaultId,
        url: file.url,
        category: "Files",
      });
    } catch (_) {}
  }
  return results;
}

async function decryptSearchFolders(folders) {
  const results = [];
  for (let i = 0; i < folders.length; i += 1) {
    const folder = folders[i];
    try {
      const opened = await vault.decryptSearchFolder(folder);
      const metadata = opened && opened.metadata ? opened.metadata : {};
      const name = opened && opened.name && opened.name.name ? opened.name.name : String(metadata.name || "Encrypted folder");
      results.push({
        id: folder.id,
        kind: "folder",
        title: name,
        meta: "Folder",
        score: Number(folder.score || 0),
        url: folder.url,
        category: "Folders",
      });
    } catch (_) {}
  }
  return results;
}

export function initSearchPalette() {
  const trigger = document.getElementById("app-search-trigger");
  const panel = document.getElementById("search-panel");
  const input = document.getElementById("search-panel-input");
  const resultsRoot = document.getElementById("search-results-items");
  const sharesRoot = document.getElementById("search-results-shares");
  const settingsRoot = document.getElementById("search-results-settings");
  const loading = document.getElementById("search-loading");
  const empty = document.getElementById("search-empty");

  if (!trigger || !panel || !input || !resultsRoot || !sharesRoot || !settingsRoot || !loading || !empty || trigger.hasAttribute("data-search-bound")) {
    return;
  }
  trigger.setAttribute("data-search-bound", "true");

  const sections = {
    results: document.querySelector("[data-search-category='results']"),
    shares: document.querySelector("[data-search-category='shares']"),
    settings: document.querySelector("[data-search-category='settings']")
  };

  let activeIndex = 0;
  let flatItems = [];
  let abortController = null;
  let searchTimer = null;
  let latestQuery = "";

  function iconFor(kind) {
    if (kind === "file") return "F";
    if (kind === "folder") return "D";
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
    item.setAttribute("data-item-id", String(result.id || ""));

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
    Array.from(root.querySelectorAll("[data-object-url]")).forEach(function(node) {
      const objectURL = node.getAttribute("data-object-url");
      if (objectURL) {
        URL.revokeObjectURL(objectURL);
      }
    });
    root.innerHTML = "";
    items.forEach(function(result) {
      root.appendChild(makeItem(result));
    });
    sectionVisibility(key, items.length);
  }

  function setLoading(isLoading) {
    loading.classList.toggle("is-hidden", !isLoading);
  }

  function mergeResults(folders, files) {
    return folders.concat(files).sort(function(a, b) {
      const scoreA = Number(a.score || 0);
      const scoreB = Number(b.score || 0);
      if (scoreA !== scoreB) return scoreB - scoreA;
      return String(a.title || "").localeCompare(String(b.title || ""));
    });
  }

  async function hydrateCachedSearchThumbnails(items) {
    for (let i = 0; i < items.length; i += 1) {
      const result = items[i];
      if (!result || result.kind !== "file") {
        continue;
      }
      const preview = result.preview || null;
      if (!preview || !preview.has_thumbnail) {
        continue;
      }
      const thumbnailVersion = Number(preview.thumbnail_version || 0);
      const thumbnailSize = Number(preview.thumbnail_size || 0);
      if (thumbnailVersion <= 0 || thumbnailSize <= 0) {
        continue;
      }
      const encryptedBytes = await thumbnailCache.get(result.id, thumbnailVersion, thumbnailSize);
      if (!(encryptedBytes instanceof Uint8Array) || !encryptedBytes.length) {
        continue;
      }
      try {
        const decrypted = await vault.decryptSearchThumbnail(result, encryptedBytes);
        const thumbBytes = decrypted && decrypted.thumbnailBytes ? decrypted.thumbnailBytes : null;
        if (!(thumbBytes instanceof Uint8Array) || !thumbBytes.length) {
          continue;
        }
        const item = resultsRoot.querySelector(".search-item[data-item-id='" + result.id + "']");
        if (!item) {
          continue;
        }
        const iconWrap = item.querySelector(".search-item-icon");
        if (!iconWrap) {
          continue;
        }
        iconWrap.textContent = "";
        iconWrap.classList.add("has-thumb");
        const objectURL = URL.createObjectURL(new Blob([thumbBytes], { type: String(preview.thumbnail_mime || "image/webp") }));
        const image = document.createElement("img");
        image.className = "search-item-thumb";
        image.alt = result.title;
        image.loading = "lazy";
        image.decoding = "async";
        image.src = objectURL;
        image.setAttribute("data-object-url", objectURL);
        iconWrap.appendChild(image);
      } catch (_) {}
    }
  }

  function renderResults(results) {
    const files = (results && results.files) || [];
    const folders = (results && results.folders) || [];
    const shares = (results && results.shares) || [];
    const settings = (results && results.settings) || [];
    const merged = mergeResults(folders, files);

    renderCategory(resultsRoot, "results", merged);
    renderCategory(sharesRoot, "shares", shares);
    renderCategory(settingsRoot, "settings", settings);

    const total = merged.length + shares.length + settings.length;
    empty.classList.toggle("is-hidden", total !== 0);
    flatten();
    hydrateCachedSearchThumbnails(merged).catch(function() {});
  }

  async function runSearch(query) {
    const settings = searchSettingsResults(query);
    latestQuery = query;
    if (abortController) {
      abortController.abort();
    }
    abortController = new AbortController();
    setLoading(true);
    try {
      const tokens = await vault.createSearchTokens(query);
      if (!tokens.length) {
        if (latestQuery === query) {
          renderResults({ folders: [], files: [], shares: [], settings: settings });
        }
        return;
      }
      const data = await apiRequest("/api/search", {
        method: "POST",
        signal: abortController.signal,
        headers: {
          "Accept": "application/json",
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ tokens: tokens, limit: 20 })
      }, {
        code: "unknown_error",
        message: "Search failed"
      });
      const folders = await decryptSearchFolders((data && data.results && data.results.folders) || []);
      const files = await decryptSearchFiles((data && data.results && data.results.files) || []);
      if (latestQuery === query) {
        renderResults({
          folders: folders,
          files: files,
          shares: [],
          settings: settings
        });
      }
    } catch (err) {
      if (err && err.name === "AbortError") return;
      if (latestQuery === query) {
        renderResults({ folders: [], files: [], shares: [], settings: settings });
      }
    } finally {
      if (latestQuery === query) {
        setLoading(false);
      }
    }
  }

  function scheduleSearch() {
    const query = input.value.trim();
    const normalized = normalizeQuery(query).replace(/\s+/g, "");
    if (searchTimer) {
      clearTimeout(searchTimer);
    }
    if (!query || normalized.length < MIN_SEARCH_QUERY_LEN) {
      latestQuery = query;
      if (abortController) {
        abortController.abort();
        abortController = null;
      }
      setLoading(false);
      renderResults({});
      return;
    }
    searchTimer = setTimeout(function() {
      searchTimer = null;
      runSearch(query);
    }, SEARCH_DEBOUNCE_MS);
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
}
