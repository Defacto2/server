/**
 * @file layout-routes.mjs
 * In-page filter for the server routes inspection page (/routes).
 */

/**
 * Initializes the in-page filter on the routes list.
 */
export default function routesFilter() {
  const init = () => {
    const input = document.getElementById('routes-filter');
    const items = document.querySelectorAll('#routes-list > .col');
    if (!input || !items.length) {
      return;
    }

    input.addEventListener('input', () => {
      const q = input.value.trim().toLowerCase();
      for (const item of items) {
        item.classList.toggle(
          'd-none',
          q !== '' && !item.textContent.toLowerCase().includes(q)
        );
      }
    });
  };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
    return;
  }
  init();
}
