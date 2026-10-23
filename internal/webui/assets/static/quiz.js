(function () {
  document.addEventListener('click', function (event) {
    const menu = document.querySelector('.quiz-new[open]');
    if (menu && !menu.contains(event.target)) menu.open = false;
  });
  document.addEventListener('keydown', function (event) {
    const menu = document.querySelector('.quiz-new[open]');
    if (menu && event.key === 'Escape') { menu.open = false; menu.querySelector('summary').focus(); }
  });
  document.addEventListener('htmx:beforeRequest', function (event) {
    if (event.detail.elt.id !== 'workspace') return;
    const catalog = document.querySelector('.quiz-catalog');
    // Polling must not replace an open menu or a keyboard selection.
    if (catalog && (catalog.querySelector('.quiz-new[open]') || catalog.contains(document.activeElement))) event.preventDefault();
  });
  function surface() { return document.querySelector('#quiz-surface'); }
  function storageKey(panel) { return 'chr-quiz-drafts:' + panel.dataset.quizContext + ':' + panel.dataset.quizSession; }
  function drafts(panel) {
    try { return Object.assign(Object.create(null), JSON.parse(sessionStorage.getItem(storageKey(panel)) || '{}')); }
    catch (_) { return {}; }
  }
  function store(panel, values) {
    try { sessionStorage.setItem(storageKey(panel), JSON.stringify(values)); }
    catch (_) { document.querySelector('#quiz-result').textContent = 'Draft recovery is unavailable. Keep this page open until your answer is recorded.'; }
  }
  function saveDraft(panel) {
    const input = panel.querySelector('#quiz-answer');
    if (!input) return;
    const values = drafts(panel);
    if (input.value) values[panel.dataset.quizQuestion] = input.value;
    else delete values[panel.dataset.quizQuestion];
    store(panel, values);
  }
  function dirty(panel) {
    const input = panel.querySelector('#quiz-answer');
    return input && input.value.length > 0 && !panel.dataset.workflowSubmitting;
  }
  function restoreMode(panel) {
    panel.querySelectorAll('#quiz-mode input').forEach(function (input) {
      input.checked = input.value === panel.dataset.quizCurrentMode;
    });
  }
  function initialStartKey(panel) { return 'chr-quiz-start:' + panel.dataset.quizContext; }
  function initialStartIntent(panel, mode) {
    const form = panel.querySelector('#quiz-mode-form');
    const revision = form.elements.revision.value;
    let previous;
    try { previous = JSON.parse(sessionStorage.getItem(initialStartKey(panel)) || 'null'); } catch (_) {}
    if (previous && previous.revision === revision && previous.mode === mode && /^web-quiz-[0-9a-f-]{36}$/.test(previous.requestID)) return previous.requestID;
    const requestID = 'web-quiz-' + window.crypto.randomUUID();
    try { sessionStorage.setItem(initialStartKey(panel), JSON.stringify({revision, mode, requestID})); } catch (_) {}
    return requestID;
  }
  function showInitialRetry(panel) {
    const retry = panel && panel.querySelector('[data-quiz-retry-start]');
    if (retry) retry.hidden = false;
  }
  function prepareFirstQuestion(panel) {
    if (panel.dataset.quizAutoStart !== 'true') return;
    const form = panel.querySelector('#quiz-mode-form');
    try {
      const previous = JSON.parse(sessionStorage.getItem(initialStartKey(panel)) || 'null');
      if (previous && previous.revision === form.elements.revision.value && ['deep', 'fast'].includes(previous.mode)) {
        form.elements.mode.value = previous.mode;
      }
    } catch (_) {}
    // Navigation selects the quiz. Keep the read request side-effect-free and
    // admit its first question through the existing guarded, retryable POST.
    if (window.htmx) window.htmx.process(form);
    setTimeout(function () { if (form.isConnected) form.requestSubmit(); }, 0);
  }
  function discard(panel) {
    const values = drafts(panel);
    delete values[panel.dataset.quizQuestion];
    store(panel, values);
    const input = panel.querySelector('#quiz-answer');
    if (input) input.value = '';
  }
  function retainedDrafts(panel, values) {
    const host = panel.querySelector('[data-quiz-retained]');
    if (!host) return;
    host.replaceChildren();
    Object.entries(values).filter(([question]) => question !== panel.dataset.quizQuestion).forEach(function ([question, text]) {
      const details = document.createElement('details');
      const summary = document.createElement('summary');
      summary.textContent = 'Unsent answer from an earlier question';
      const body = document.createElement('p');
      body.textContent = text;
      const button = document.createElement('button');
      button.type = 'button'; button.className = 'button'; button.textContent = 'Discard this draft';
      button.addEventListener('click', function () {
        if (!window.confirm('Discard this unsent answer?')) return;
        const latest = drafts(panel); delete latest[question]; store(panel, latest); details.remove();
      });
      details.append(summary, body, button); host.append(details);
    });
  }
  function setup() {
    const panel = surface();
    if (!panel) return;
    if (panel.dataset.quizBound) {
      const failure = panel.querySelector('[data-quiz-write-status]');
      if (failure && failure.dataset.quizWriteStatus !== 'ok' && panel.dataset.quizSession) restoreMode(panel);
      if (failure && failure.dataset.quizWriteStatus !== 'ok') showInitialRetry(panel);
      return;
    }
    panel.dataset.quizBound = '1';
    const values = drafts(panel);
    if (panel.dataset.ackAction === 'answer' && panel.dataset.ackSession === panel.dataset.quizSession) {
      delete values[panel.dataset.ackQuestion]; store(panel, values);
    }
    if (panel.dataset.ackAction === 'start') {
      try { sessionStorage.removeItem(initialStartKey(panel)); } catch (_) {}
    }
    const input = panel.querySelector('#quiz-answer');
    if (input) {
      input.value = values[panel.dataset.quizQuestion] || '';
      input.addEventListener('input', function () { saveDraft(panel); });
    }
    retainedDrafts(panel, values);
    const modeForm = panel.querySelector('[data-quiz-mode="true"]');
    if (modeForm) modeForm.addEventListener('change', function (event) {
      if (event.target.matches('[name="mode"]')) modeForm.requestSubmit();
    });
    if (panel.dataset.ackAction === 'mode') panel.querySelector('#quiz-mode input:checked').focus({preventScroll: true});
    prepareFirstQuestion(panel);
  }
  document.addEventListener('submit', function (event) {
    const form = event.target.closest('[data-quiz-form]');
    const panel = surface();
    if (!form || !panel) return;
    const action = event.submitter && event.submitter.value || form.elements.action.value;
    if (['quiz.skip', 'quiz.cancel', 'quiz.done'].includes(action) && dirty(panel)) {
      if (!window.confirm('Discard this unsent answer and ' + action.slice(5) + '?')) {
        event.preventDefault(); event.stopImmediatePropagation(); return;
      }
      discard(panel);
    }
    saveDraft(panel);
    form.dataset.quizAction = action;
    const fingerprint = JSON.stringify([action, form.elements.revision.value, form.elements.answer && form.elements.answer.value, form.elements.mode && form.elements.mode.value]);
    if (form.dataset.quizArguments !== fingerprint) {
      form.elements.request_id.value = action === 'quiz.start' && panel.dataset.quizAutoStart === 'true' ? initialStartIntent(panel, form.elements.mode.value) : 'web-quiz-' + window.crypto.randomUUID();
      form.dataset.quizArguments = fingerprint;
    }
  }, true);
  document.addEventListener('htmx:beforeRequest', function (event) {
    const form = event.detail.elt.closest('[data-quiz-form]');
    const panel = surface();
    if (!form || !panel) return;
    const action = form.dataset.quizAction;
    const retry = panel.querySelector('[data-quiz-retry-start]');
    if (retry) retry.hidden = true;
    const messages = {'quiz.start': 'Preparing your first question…', 'quiz.answer': 'Saving your answer and preparing the next question…', 'quiz.skip': 'Preparing the next question…', 'quiz.mode': 'Switching question mode…', 'quiz.resume': 'Resuming your quiz…', 'quiz.done': 'Finishing your quiz…', 'quiz.cancel': 'Saving your progress…'};
    panel.querySelector('.quiz-progress').textContent = messages[action] || 'Updating your quiz…';
    panel.querySelector('.quiz-panel').setAttribute('aria-busy', 'true');
  });
  document.addEventListener('htmx:afterRequest', function (event) {
    if (!event.detail.elt.closest('[data-quiz-form]')) return;
    const panel = surface();
    if (!panel) return;
    panel.querySelector('.quiz-progress').textContent = '';
    panel.querySelector('.quiz-panel').removeAttribute('aria-busy');
  });
  document.addEventListener('click', function (event) {
    const link = event.target.closest('a[href]');
    const panel = surface();
    if (!link || !panel || link.matches('[data-quiz-reload]') || !dirty(panel)) return;
    if (!window.confirm('Discard this unsent answer and leave the quiz?')) { event.preventDefault(); event.stopImmediatePropagation(); return; }
    discard(panel);
  }, true);
  window.addEventListener('beforeunload', function (event) {
    const panel = surface();
    if (!panel || !dirty(panel)) return;
    saveDraft(panel); event.preventDefault(); event.returnValue = '';
  });
  ['htmx:sendError', 'htmx:responseError', 'htmx:timeout'].forEach(function (name) {
    document.addEventListener(name, function (event) {
      const form = event.detail.elt.closest('[data-quiz-form]');
      if (!form) return;
      const panel = surface(); if (panel) saveDraft(panel);
      showInitialRetry(panel);
      const result = document.querySelector('#quiz-result');
      result.textContent = 'The request failed. Your draft is retained. Retry the same action or reload the current question. ';
      const reload = document.createElement('a'); reload.href = location.href; reload.dataset.quizReload = ''; reload.textContent = 'Reload current question'; result.append(reload);
      if (panel && panel.dataset.quizSession) restoreMode(panel);
    });
  });
  document.addEventListener('DOMContentLoaded', setup);
  document.addEventListener('htmx:afterSwap', setup);
})();
