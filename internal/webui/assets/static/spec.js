(function () {
  document.addEventListener('submit', function (event) {
    const form = event.target.closest('[data-spec-form]');
    if (!form) return;
    const args = JSON.stringify(Array.from(new FormData(form)).filter(([key]) => key !== 'request_id'));
    if (form.dataset.specArguments && form.dataset.specArguments !== args) form.elements.request_id.value = 'web-spec-' + crypto.randomUUID();
    form.dataset.specArguments = args;
  }, true);
  ['htmx:responseError', 'htmx:sendError', 'htmx:timeout'].forEach(function (name) {
    document.addEventListener(name, function (event) {
      const form = event.detail.elt.closest('[data-spec-form]');
      if (form) form.querySelector('#spec-result').textContent = 'The request failed. Your edits are retained. Retry Create or Save.';
    });
  });
})();
