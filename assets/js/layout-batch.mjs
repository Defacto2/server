/**
 * @file layout-batch.mjs
 * Editor helper to open artifact cards in batches of browser tabs.
 */

const BATCH_SIZE = 25;

/**
 * Populates the editor "Open in tabs" dropdown on artifact listing pages.
 */
export default function batchOpenTabs() {
  const init = () => {
    const menu = document.getElementById('batch-open-menu');
    const dropdown = document.getElementById('batch-open-dropdown');
    if (!menu || !dropdown) {
      return;
    }

    const urls = Array.from(
      document.querySelectorAll('[data-artifact-url]'),
      (el) => el.dataset.artifactUrl
    ).filter(Boolean);
    if (urls.length === 0) {
      dropdown.classList.add('d-none');
      return;
    }

    const storageKey = `batch-open:${window.location.pathname}`;
    const opened = loadOpened(storageKey);
    menu.textContent = '';

    for (let start = 0; start < urls.length; start += BATCH_SIZE) {
      const end = Math.min(start + BATCH_SIZE, urls.length);
      const li = document.createElement('li');
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.className =
        'dropdown-item d-flex justify-content-between align-items-center gap-3';

      const label = document.createElement('span');
      label.textContent = `${start + 1}\u2013${end}`;
      btn.appendChild(label);

      const badge = document.createElement('span');
      badge.className = opened.has(start)
        ? 'badge text-bg-secondary'
        : 'badge text-bg-secondary d-none';
      badge.textContent = 'opened';
      btn.appendChild(badge);

      btn.addEventListener('click', () => {
        const openedCount = openRange(urls, start, end);
        if (openedCount > 0) {
          opened.add(start);
          saveOpened(storageKey, opened);
          badge.classList.remove('d-none');
        }
        updateAlert(openedCount, end - start);
      });

      li.appendChild(btn);
      menu.appendChild(li);
    }
  };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
    return;
  }
  init();
}

/**
 * Synchronously opens a slice of URLs in new tabs inside a user click event.
 * @param {string[]} urls
 * @param {number} start
 * @param {number} end
 * @returns {number} Number of tabs opened before completion or pop-up block.
 */
function openRange(urls, start, end) {
  let count = 0;
  for (let i = start; i < end; i++) {
    const win = window.open(urls[i], '_blank');
    if (!win || win.closed || typeof win.closed === 'undefined') {
      break;
    }
    count++;
  }
  return count;
}

/**
 * Displays or hides the pop-up blocker warning alert.
 * @param {number} openedCount
 * @param {number} expectedCount
 */
function updateAlert(openedCount, expectedCount) {
  const alertEl = document.getElementById('batch-open-alert');
  if (!alertEl) {
    return;
  }
  if (openedCount < expectedCount) {
    alertEl.textContent = `Opened ${openedCount} of ${expectedCount} tabs before the browser blocked pop-ups. Allow pop-ups for this site and try again.`;
    alertEl.classList.remove('d-none');
    return;
  }
  alertEl.textContent = '';
  alertEl.classList.add('d-none');
}

/**
 * Loads previously opened batch start indices from sessionStorage.
 * @param {string} key
 * @returns {Set<number>}
 */
function loadOpened(key) {
  try {
    const raw = window.sessionStorage.getItem(key);
    if (!raw) {
      return new Set();
    }
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? new Set(parsed) : new Set();
  } catch {
    return new Set();
  }
}

/**
 * Saves opened batch start indices to sessionStorage.
 * @param {string} key
 * @param {Set<number>} opened
 */
function saveOpened(key, opened) {
  try {
    window.sessionStorage.setItem(key, JSON.stringify([...opened]));
  } catch {
    // Ignore storage quota or privacy mode errors.
  }
}
