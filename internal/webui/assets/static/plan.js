(function () {
  function prepare(form) {
    const selected = form.elements.case_slug.selectedOptions[0];
    if (!selected) return;
    form.elements.id.value = selected.value;
    form.elements.revision.value = selected.dataset.planRevision;
    form.querySelector('[data-plan-spec]').textContent = 'Spec: ' + selected.dataset.specTitle;
  }
  document.addEventListener('change', function (event) {
    const form = event.target.closest('[data-plan-generate]');
    if (!form) return;
    form.elements.request_id.value = '';
    prepare(form);
  });
  document.addEventListener('htmx:afterSwap', function () {
    document.querySelectorAll('[data-plan-generate]').forEach(prepare);
  });
  document.addEventListener('click', function (event) {
    if (event.target.closest('[data-modal-template="plan-generate-modal"]')) {
      document.querySelectorAll('[data-plan-generate]').forEach(prepare);
    }
  });
  document.addEventListener('submit', function (event) {
    const form = event.target.closest('[data-plan-generate]');
    if (!form) return;
    prepare(form);
    if (!form.elements.request_id.value) {
      const bytes = window.crypto.getRandomValues(new Uint8Array(16));
      form.elements.request_id.value = 'web-plan-' + Array.from(bytes, byte => byte.toString(16).padStart(2, '0')).join('');
    }
  }, true);
  document.addEventListener('htmx:beforeRequest', function (event) {
    const form = event.detail.elt.closest('[data-plan-generate]');
    if (form) form.querySelector('#plan-result').textContent = 'Generating plan and tickets…';
  });
  ['htmx:responseError', 'htmx:sendError', 'htmx:timeout'].forEach(function (name) {
    document.addEventListener(name, function (event) {
      const form = event.detail.elt.closest('[data-plan-generate]');
      if (form) form.querySelector('#plan-result').textContent = 'The request failed. Retry Generate Plan to continue the same operation.';
    });
  });
})();
