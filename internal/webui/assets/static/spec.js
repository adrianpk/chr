(function () {
  function newRequestID() {
    const bytes = new Uint8Array(16);
    window.crypto.getRandomValues(bytes);
    return 'web-spec-' + Array.from(bytes, byte => byte.toString(16).padStart(2, '0')).join('');
  }
  document.addEventListener('submit', function (event) {
    const form = event.target.closest('[data-spec-form]');
    if (!form) return;
    const args = JSON.stringify(Array.from(new FormData(form)).filter(([key]) => key !== 'request_id'));
    if (!form.elements.request_id.value || (form.dataset.specArguments && form.dataset.specArguments !== args)) form.elements.request_id.value = newRequestID();
    form.dataset.specArguments = args;
  }, true);
  document.addEventListener('htmx:beforeRequest', function (event) {
    const form = event.detail.elt.closest('[data-spec-generate]');
    if (form) form.querySelector('#spec-result').textContent = 'Generating specification…';
  });
  ['htmx:responseError', 'htmx:sendError', 'htmx:timeout'].forEach(function (name) {
    document.addEventListener(name, function (event) {
      const form = event.detail.elt.closest('[data-spec-form]');
      if (form) form.querySelector('#spec-result').textContent = form.matches('[data-spec-generate]') ? 'The request failed. Your selected scope is retained. Retry Generate.' : 'The request failed. Your edits are retained. Retry Create or Save.';
    });
  });
})();
